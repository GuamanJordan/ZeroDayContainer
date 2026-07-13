# Plan de Implementación: Contenedor desde Cero (1 mes)

**Proyecto:** Runtime de contenedores ligero en Go, sin Docker.
**Duración:** 4 semanas, 5 días hábiles por semana (20 días de trabajo + días de buffer/repaso).
**Dedicación estimada:** 2-3 horas/día.
**Repo sugerido:** `micro-container/` con estructura modular por componente (namespaces, cgroups, network, cli).

> Nota sobre entorno: necesitas una VM Linux con kernel ≥ 5.x y acceso root (WSL2 con systemd, una VM en VirtualBox/UTM, o una instancia cloud). Los namespaces, cgroups y mount requieren privilegios; **no lo desarrolles directamente en el host de producción**.

---

## Semana 0 (previa, medio día): Setup del entorno

- [ ] Instalar Go 1.22+, `strace`, `iproute2`, `util-linux` (para `unshare`, `nsenter`, `setarch`).
- [ ] Descargar un rootfs mínimo: `alpine-minirootfs` (busybox + apk).
- [ ] Crear repo en GitHub con estructura inicial:
  ```
  micro-container/
    cmd/mc/main.go
    internal/namespaces/
    internal/cgroups/
    internal/network/
    internal/rootfs/
    test/
    .github/workflows/
    Makefile
  ```
- [ ] Configurar `golangci-lint` y `gosec` localmente.
- [ ] Primer commit + primer workflow de CI vacío (solo `go build`).

---

## Semana 1 — Fundamentos: procesos, namespaces básicos y CI inicial

**Objetivo de la semana:** entender `fork`/`clone`/`unshare`, crear un proceso aislado con PID, UTS y mount namespace, y tener un pipeline de CI corriendo tests unitarios en cada push.

### Día 1 — Fork, exec y el árbol de procesos
**Concepto:** en Linux todo proceso nuevo nace de `fork()` (copia el proceso padre) seguido normalmente de `exec()` (reemplaza la imagen de memoria por un nuevo programa). Entender esto es la base de por qué `clone()` (que usan los contenedores) es "fork con esteroides": permite especificar *qué* se comparte con el hijo (memoria, namespaces, etc.) mediante flags.
**Guía:** [Día 1: fork, clone, exec y árbol de procesos](../dia-01/README.md)
**Actividades:**
- Leer `man fork`, `man execve`, `man clone`.
- Escribir un programa pequeño en Go usando `os/exec` que lance un subproceso y capture su PID.
- Usar `strace -f` sobre un comando simple (`ls`) para ver la secuencia `clone → execve → exit`.
**Entregable:** notas cortas (README) explicando fork vs clone vs exec con tus palabras.

### Día 2 — Namespaces: teoría y exploración manual
**Concepto:** un namespace virtualiza un recurso global del kernel para que un grupo de procesos vea "su propia versión" de él. Los 7 tipos: PID, network, mount, UTS, IPC, user, cgroup.
**Actividades:**
- Reproducir en tu VM los comandos del artículo de referencia (namespace de red con `unshare -n`, veth pair).
- Inspeccionar `/proc/$$/ns/` antes y después de `unshare`.
- Probar `unshare --pid --fork --mount-proc` y observar que `ps` solo ve el nuevo árbol.
**Entregable:** captura/log de terminal mostrando namespaces distintos para dos procesos.

### Día 3 — `clone()` desde Go: CLONE_NEWPID y CLONE_NEWUTS
**Concepto:** en Go se accede a `clone()` vía `syscall.SysProcAttr{Cloneflags: ...}` al lanzar un `exec.Cmd`. Con `CLONE_NEWUTS` el hijo puede tener su propio hostname; con `CLONE_NEWPID` el hijo se convierte en PID 1 de su propio árbol.
**Actividades:**
- Escribir `internal/namespaces/run.go`: función que lanza un proceso hijo con `CLONE_NEWUTS|CLONE_NEWPID`.
- Dentro del hijo, cambiar el hostname (`sethostname`) y verificar que el host no lo ve afectado.
- Verificar con `echo $$` dentro del namespace que el PID es 1.
**Entregable:** comando `mc run-basic <cmd>` funcional que aísla UTS+PID.

### Día 4 — Mount namespace y montaje de `/proc`
**Concepto:** sin un mount namespace propio y sin remontar `/proc`, comandos como `ps` dentro del contenedor seguirán mostrando los procesos del host (porque `/proc` es una vista, no algo mágicamente aislado). Hay que añadir `CLONE_NEWNS` y montar un `/proc` nuevo dentro.
**Actividades:**
- Añadir `CLONE_NEWNS` a las flags.
- Montar `proc` con `mount("proc", "/proc", "proc", 0, "")` dentro del hijo.
- Verificar que `ps aux` dentro del namespace solo muestra sus propios procesos.
**Entregable:** `mc run-basic` ahora también aísla mount, con `/proc` propio funcionando.

### Día 5 — Tests unitarios + primer pipeline de CI real
**Concepto de CI:** en este proyecto conviene separar **tests unitarios** (lógica pura, sin privilegios, corren en cualquier runner) de **tests de integración** (requieren root/namespaces, corren en un job con `privileged` o en un runner self-hosted). Hoy armamos la primera capa.
**Actividades:**
- Escribir tests unitarios para parseo de flags y construcción de `SysProcAttr` (sin ejecutar namespaces reales, solo validar la config generada).
- Crear `.github/workflows/ci.yml` con job `unit-tests`:
  ```yaml
  name: CI
  on: [push, pull_request]
  jobs:
    unit-tests:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.22' }
        - run: go build ./...
        - run: go vet ./...
        - run: go test ./... -v -race -cover
  ```
- Añadir `golangci-lint` como job separado.
**Entregable:** badge de CI verde en el README.

**Repaso fin de semana 1:** día de buffer para releer `namespaces(7)` completo y limpiar el código de la semana.

---

## Semana 2 — Filesystem: chroot, pivot_root y rootfs

**Objetivo:** aislar el sistema de archivos que ve el proceso contenedor, y empezar a correr tests de integración reales en CI usando contenedores privilegiados (Docker-in-Docker).

### Día 6 — chroot: la abuela del aislamiento de filesystem
**Concepto:** `chroot()` cambia la raíz aparente `/` de un proceso, pero **no es seguro** por sí solo (existen escapes clásicos vía descriptores de archivo abiertos antes del chroot, o `..` mal manejado). Aun así, entenderlo es clave antes de pasar a `pivot_root`.
**Actividades:**
- Descomprimir el rootfs de Alpine en `/opt/rootfs`.
- Desde Go, llamar a `syscall.Chroot("/opt/rootfs")` seguido de `os.Chdir("/")`.
- Ejecutar `/bin/sh` dentro y confirmar que no se ve el filesystem del host.
- Investigar y documentar (sin implementar) un escape clásico de chroot, para entender por qué la industria migró a `pivot_root`.
**Entregable:** demo de chroot + doc de 1 párrafo sobre sus limitaciones de seguridad.

### Día 7 — pivot_root: la forma "correcta"
**Concepto:** `pivot_root(new_root, put_old)` mueve el root filesystem actual a un directorio dentro del nuevo root y hace que `new_root` sea el `/` real, permitiendo además desmontar el root viejo. Requiere que `new_root` sea un punto de montaje (bind mount de sí mismo si hace falta).
**Actividades:**
- Implementar `internal/rootfs/pivot.go`.
- Secuencia: bind-mount del rootfs sobre sí mismo → crear `.old_root` dentro → `pivot_root` → `chdir("/")` → `umount(".old_root")` con `MNT_DETACH` → `rmdir(".old_root")`.
- Probar y comparar contra el chroot del día anterior con `strace`.
**Entregable:** `mc run` usa `pivot_root` en vez de `chroot`.

### Día 8 — Montajes esenciales dentro del contenedor
**Concepto:** un rootfs "vacío" necesita `/proc`, `/sys`, `/dev` (al menos `/dev/null`, `/dev/zero`, `/dev/tty`) para que binarios comunes funcionen. Esto se hace con mounts propios post-pivot.
**Actividades:**
- Montar `proc`, `sysfs`, y crear un `tmpfs` mínimo en `/dev` con los nodos básicos (`mknod` para null, zero, tty, random).
- Probar que `busybox` dentro del contenedor puede correr `ls`, `ps`, `cat /proc/cpuinfo` sin errores.
**Entregable:** contenedor funcional que ejecuta un shell interactivo con filesystem y /proc propios.

### Día 9 — Tests de integración con privilegios en CI
**Concepto de CI:** GitHub Actions `ubuntu-latest` runners permiten `sudo` y namespaces sin problema (no necesitas Docker-in-Docker para esto, a diferencia de otros CI). Aprovechamos eso para correr los tests reales de namespaces/mounts directamente.
**Actividades:**
- Crear job `integration-tests` en el workflow, separado de `unit-tests`:
  ```yaml
    integration-tests:
      runs-on: ubuntu-latest
      needs: unit-tests
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.22' }
        - run: go build -o bin/mc ./cmd/mc
        - name: Run integration tests (requires root)
          run: sudo go test ./test/integration/... -tags=integration -v
  ```
- Escribir un test de integración: levantar el contenedor, ejecutar `hostname` y `ps`, verificar aislamiento por output.
- Usar build tag `//go:build integration` para separarlos de los unitarios.
**Entregable:** pipeline con dos etapas (unit + integration) corriendo en cada PR.

### Día 10 — Buffer / hardening de filesystem
**Actividades:**
- Manejar rootfs de solo lectura opcional (`MS_RDONLY`).
- Limpiar errores, agregar manejo de `defer` para desmontajes en caso de fallo.
- Documentar en README el flujo completo de pivot_root con diagrama (puedes pedírmelo aparte si quieres el diagrama visual).

---

## Semana 3 — cgroups, capabilities y red

**Objetivo:** limitar recursos (CPU/memoria), reducir privilegios, y dar red aislada con salida a internet.

### Día 11 — cgroups v2: teoría
**Concepto:** los cgroups (control groups) limitan y contabilizan uso de recursos (CPU, memoria, I/O, PIDs) de un grupo de procesos. A diferencia de namespaces, no hay syscalls especiales: es leer/escribir archivos de texto en `/sys/fs/cgroup/`. Verifica primero si tu sistema corre cgroups v2 (`cat /sys/fs/cgroup/cgroup.controllers`).
**Actividades:**
- Explorar manualmente: crear un cgroup con `mkdir /sys/fs/cgroup/mc-test`, escribir límites en `memory.max` y `cpu.max`.
- Lanzar un proceso que consuma memoria de más (`stress-ng` o un script simple) y verificar que el OOM killer del cgroup lo mata.
**Entregable:** demo manual de límite de memoria funcionando.

### Día 12 — cgroups desde Go
**Actividades:**
- Implementar `internal/cgroups/cgroup.go`: crear el cgroup, escribir `memory.max`, `cpu.max`, `pids.max`, y mover el PID del proceso contenedor a `cgroup.procs`.
- Integrar con flags de CLI: `--memory=100m --cpus=0.5`.
- Test de integración: lanzar un proceso que intente exceder el límite de memoria y verificar que es matado.
**Entregable:** `mc run --memory=50m <cmd>` respeta el límite.

### Día 13 — Capabilities: dejar de correr como root pleno
**Concepto:** Linux capabilities dividen los privilegios de root en unidades granulares (`CAP_NET_ADMIN`, `CAP_SYS_ADMIN`, etc.). Un contenedor "real" no debería tener todas las capabilities del root del host; se dropean las innecesarias para reducir superficie de ataque.
**Actividades:**
- Investigar `libcap` y el paquete Go `github.com/syndtr/gocapability` (o llamadas directas a `prctl`/`capset`).
- Dropear capabilities peligrosas (`CAP_SYS_ADMIN` fuera de lo estrictamente necesario, `CAP_SYS_MODULE`, etc.) antes del `execve` final.
**Entregable:** el proceso final corre con un set reducido de capabilities, verificable con `getpcaps <pid>`.

### Día 14 — Network namespace + veth pair
**Concepto:** aplicar exactamente lo que viste en el artículo de referencia: crear un network namespace propio para el contenedor, un par veth, asignar IPs, y habilitar tráfico entre host y contenedor.
**Actividades:**
- Automatizar en Go lo que hiciste manualmente con `ip` en la Semana 1: `CLONE_NEWNET`, crear veth pair vía netlink (paquete `github.com/vishvananda/netlink` recomendado) o *shelling out* a `ip` como primera versión simple.
- Asignar `10.16.8.1/24` al host y `10.16.8.2/24` al contenedor, subir ambas interfaces.
- Verificar ping bidireccional.
**Entregable:** `internal/network/veth.go` funcional.

### Día 15 — Bridge + NAT para salida a internet (opcional pero recomendado)
**Concepto:** para que el contenedor tenga acceso real a internet (no solo al host), se necesita un bridge en el host conectando varios veth, más reglas NAT (`iptables -t nat -A POSTROUTING -j MASQUERADE`) que traduzcan la IP interna a la IP pública del host.
**Actividades:**
- Crear bridge `mc0`, conectar el veth del host a él.
- Configurar `iptables` para NAT y habilitar `ip_forward` (`sysctl net.ipv4.ip_forward=1`).
- Probar `ping 8.8.8.8` desde dentro del contenedor.
**Entregable:** contenedor con salida a internet real.

**Nota de CI para esta semana:** los tests de red e iptables son más frágiles en runners compartidos (pueden faltar módulos de kernel o permisos NAT). Márcalos con un build tag distinto (`//go:build network_integration`) y córrelos condicionalmente, documentando en el README cómo correrlos localmente si CI no los soporta completos.

---

## Semana 4 — CLI, integración final, seguridad y pipeline completo

**Objetivo:** unir todos los componentes en una CLI coherente, pulir el pipeline de CI/CD, y (opcional) dar soporte a imágenes por capas con overlayfs.

### Día 16 — Diseño de la CLI
**Actividades:**
- Definir subcomandos: `mc run [flags] <rootfs> <cmd>`, `mc list`, `mc exec <id> <cmd>` (usando `setns` para entrar a un namespace existente).
- Usar `cobra` o flags nativos de Go para parsear argumentos.
- Refactor: mover la lógica de namespaces/cgroups/network detrás de una interfaz `Container` clara.
**Entregable:** CLI usable con `--help` decente.

### Día 17 — `setns`: entrar a un contenedor corriendo
**Concepto:** `setns(fd, nstype)` permite que un proceso existente se una a un namespace ya creado (referenciado por su fd en `/proc/<pid>/ns/*`). Es lo que usa `nsenter`/`docker exec`.
**Actividades:**
- Implementar `mc exec <pid> <cmd>`: abre `/proc/<pid>/ns/{pid,mnt,net,uts}` y hace `setns` antes de `exec`.
- Test de integración: lanzar contenedor de larga duración, ejecutar `mc exec` y verificar que ve el mismo hostname/procesos.
**Entregable:** `mc exec` funcional.

### Día 18 — overlayfs para "imágenes" por capas (opcional, nivel Docker real)
**Concepto:** overlayfs combina un `lowerdir` (capa base, solo lectura) con un `upperdir` (cambios del contenedor) y un `workdir`, presentando una vista unificada (`merged`). Es la base de cómo Docker permite reusar capas de imagen entre contenedores.
**Actividades (si el tiempo lo permite):**
- Montar overlayfs manualmente con `mount -t overlay overlay -o lowerdir=...,upperdir=...,workdir=... merged/`.
- Integrarlo como paso previo al `pivot_root`, reemplazando el rootfs plano.
**Entregable:** dos contenedores compartiendo la misma capa base sin duplicar el rootfs en disco.

### Día 19 — Seguridad y hardening final
**Actividades:**
- Añadir `gosec` como job de CI (análisis estático de seguridad en Go).
- Revisar: ¿el rootfs se monta con `nosuid,nodev`? ¿se dropean todas las capabilities no usadas? ¿`no_new_privs` está seteado vía `prctl`?
- Escribir un `THREAT_MODEL.md` corto: qué aísla este runtime y qué NO garantiza (esto es honestidad técnica clave, ya que un runtime educativo no es tan seguro como runc/gVisor).
**Entregable:** documento de modelo de amenazas + CI con seguridad estática.

### Día 20 — Pipeline de CI/CD completo y release
**Concepto de CI/CD:** hasta ahora solo teníamos "integración continua" (build+test en cada push). Hoy cerramos con "entrega continua": generar binarios listos para usar en cada tag de versión.
**Actividades:**
- Workflow final con 4 jobs: `lint` → `unit-tests` → `integration-tests` → `build-release` (este último solo en tags `v*`, usando `goreleaser` o `go build` cross-compilado para linux/amd64 y linux/arm64).
- Añadir badges de cobertura (Codecov opcional) y de build al README.
- Escribir el README final: instalación, uso, arquitectura, limitaciones, roadmap futuro (namespaces de usuario, seccomp, etc.).
**Entregable:** tag `v0.1.0` con binario generado automáticamente por CI.

---

## Resumen del pipeline de CI final

```yaml
name: CI/CD
on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go install github.com/golangci-lint/cmd/golangci-lint@latest
      - run: golangci-lint run
      - run: go install github.com/securego/gosec/v2/cmd/gosec@latest
      - run: gosec ./...

  unit-tests:
    needs: lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test ./... -race -cover

  integration-tests:
    needs: unit-tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go build -o bin/mc ./cmd/mc
      - run: sudo go test ./test/integration/... -tags=integration -v

  build-release:
    needs: integration-tests
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: GOOS=linux GOARCH=amd64 go build -o bin/mc-linux-amd64 ./cmd/mc
      - run: GOOS=linux GOARCH=arm64 go build -o bin/mc-linux-arm64 ./cmd/mc
      - uses: softprops/action-gh-release@v2
        with:
          files: bin/*
```

---

## Checklist de conceptos que debes poder explicar al final del mes

- [ ] Diferencia entre `fork`, `clone` y `exec`.
- [ ] Los 7 namespaces de Linux y qué aísla cada uno.
- [ ] Por qué `chroot` no es seguro y qué resuelve `pivot_root`.
- [ ] Cómo funcionan cgroups v2 (jerarquía de archivos, no syscalls).
- [ ] Qué son las capabilities y por qué "dropearlas" importa.
- [ ] Cómo se conectan dos network namespaces con veth + bridge + NAT.
- [ ] Qué hace `setns` y cómo `docker exec` se apoya en él.
- [ ] Qué es overlayfs y cómo permite compartir capas entre contenedores.
- [ ] Diferencia entre integración continua y entrega continua en este proyecto.
- [ ] Modelo de amenazas de tu propio runtime: qué SÍ y qué NO aísla comparado con Docker/runc real.
