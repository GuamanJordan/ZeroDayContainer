# ZeroDayContainer (`mc`)

[![CI](https://github.com/GuamanJordan/ZeroDayContainer/actions/workflows/ci.yml/badge.svg)](https://github.com/GuamanJordan/ZeroDayContainer/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.22%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Runtime de contenedores ligero escrito en Go desde cero, diseñado para aislar procesos en Linux utilizando **Namespaces**, **Cgroups v2**, **Capabilities** y **OverlayFS**, sin depender de Docker ni `runc`.

---

## 🚀 Características y Roadmap

Este proyecto sigue una hoja de ruta estructurada de 30 días ([Plan Completo de Implementación](docs/plan/plan-contenedor-un-mes.md)):

- [x] **Día 01:** Análisis de `fork`, `execve` y `clone`.
- [x] **Día 02:** Exploración manual de los 7+1 Namespaces de Linux.
- [x] **Día 03:** Aislamiento de UTS (hostname) y PID (proceso PID 1).
- [x] **Día 04:** Aislamiento de Mount namespace (`CLONE_NEWNS`) y `/proc` privado.
- [x] **Día 05:** CI Workflow, pruebas unitarias y linting.
- [x] **Día 06:** Aislamiento del sistema de archivos con `chroot`.
- [x] **Día 07:** Reemplazo de la raíz en contenedores con `pivot_root`.
- [x] **Día 08:** Montajes de pseudo-sistemas de archivos esenciales (`/proc`, `/sys`, `/dev`, `/dev/pts`).
- [x] **Día 09:** Pruebas de integración privilegiadas y job en CI.
- [x] **Día 10:** Hardening de filesystem (rootfs read-only opcional, desmontajes seguros y rollback).
- [x] **Día 11:** Teoría y exploración manual de Cgroups v2.
- [x] **Día 12:** Control de recursos (memoria, CPU, PIDs) con Cgroups v2 desde Go.
- [x] **Día 13:** Reducción de Linux Capabilities y activación de `no_new_privs`.
- [x] **Día 14:** Network namespace, veth pair y conectividad host-contenedor.
- [x] **Día 15:** Conexión de bridge `mc0`, enrutamiento NAT y salida a internet.
- [x] **Día 16:** Diseño modular de la CLI y persistencia de estado de contenedores (`mc list`, `mc version`).
- [x] **Día 17:** Inspección y debug con `setns(2)` y `nsenter` (`mc exec`).
- [x] **Día 18:** Sistemas de archivos en capas con OverlayFS (`--overlay`).
- [x] **Día 19:** Modelo de amenazas formal ([STRIDE / Threat Model](docs/THREAT_MODEL.md)) y escaneo con Gosec.
- [ ] **Próximos días:** Release v0.1.0 y distribución de artefactos.

---

## 🛠️ Uso de la CLI (`mc`)

### Requisitos
- Linux con kernel ≥ 5.x y privilegios de `root` (o WSL2 con systemd).
- Go 1.22+.

### Compilación

```bash
go build -o mc ./cmd/mc
```

### Ejecutar un contenedor completo con `pivot_root` (`mc run`)

```bash
# Modo estándar (lectura y escritura)
sudo ./mc run /tmp/alpine-rootfs /bin/sh

# Con capa de unión OverlayFS efímera (--overlay)
sudo ./mc run --overlay /tmp/alpine-rootfs /bin/sh

# Modo seguro con rootfs de solo lectura (--read-only)
sudo ./mc run --read-only /tmp/alpine-rootfs /bin/sh

# Con límites de recursos (cgroups v2: memoria, CPU, PIDs)
sudo ./mc run --memory=100m --cpus=0.5 --pids=30 /tmp/alpine-rootfs /bin/sh

# Con pila de red aislada y veth pair (--net)
sudo ./mc run --net /tmp/alpine-rootfs /bin/sh

# Con conexión a internet vía bridge mc0 y NAT (--nat)
sudo ./mc run --nat /tmp/alpine-rootfs /bin/sh
```

### Ejecutar comandos dentro de un contenedor activo (`mc exec`)

```bash
# Ejecutar un shell interactivo por Container ID
sudo ./mc exec mc-123456 /bin/sh

# Ejecutar un comando por PID directo
sudo ./mc exec 14209 /bin/ps aux
```

### Listar contenedores registrados (`mc list` o `mc ps`)

```bash
./mc list
```

### Consultar versión del runtime

```bash
./mc version
```

### Ejecutar subproceso básico aislado (UTS, PID, Mount)

```bash
sudo ./mc run-basic /bin/bash
```

---

## 🧪 Pruebas y Calidad de Código

### Pruebas Unitarias y Linters
```bash
go vet ./...
go test ./... -v -race -cover
```

### Pruebas de Integración (Privilegiadas)
```bash
sudo go test ./test/integration/... -tags=integration -v
```

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.
