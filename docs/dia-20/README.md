# Día 20: Release v0.1.0 y Cierre del Roadmap de 20 Días

## Objetivo

Completar el ciclo de desarrollo de 20 días de **ZeroDayContainer** (`mc`), empaquetando la versión oficial **v0.1.0** con soporte de compilación cruzada multiarquitectura (`linux/amd64` y `linux/arm64`), generación de sumas de comprobación criptográficas (**SHA-256**) y publicación automatizada de artefactos en GitHub Releases.

Al finalizar este día debes poder explicar:

- Cómo se integran de manera armónica todas las primitivas del kernel Linux (Namespaces, Cgroups v2, Capabilities, PivotRoot, OverlayFS y Red virtual) para dar vida a un contenedor funcional sin Docker ni runc.
- La estrategia de compilación cruzada estática (`CGO_ENABLED=0`) y stripping de símbolos (`-ldflags="-s -w"`).
- La verificación de integridad criptográfica de binarios distribuidos mediante `sha256sum`.
- El flujo de release continuo activado por etiquetas de versión semántica (`v*`).

---

## 🏆 Resumen Integral del Ciclo de 20 Días

```text
                               ZeroDayContainer (mc v0.1.0)
+---------------------------------------------------------------------------------------+
|                                    CLI MODULAR (mc)                                   |
|       mc run [--read-only] [--overlay] [--net] [--nat] [--memory] [--cpus] [--pids]   |
|       mc exec <pid_or_id> <cmd>    |    mc list / ps    |    mc version               |
+---------------------------------------------------------------------------------------+
        |                             |                         |
        v                             v                         v
+-----------------------+   +-------------------+   +-----------------------+
|       SEGURIDAD       |   |  ALMACENAMIENTO   |   |          RED          |
|  - Drop Capabilities  |   |  - pivot_root     |   |  - Network Namespace  |
|  - no_new_privs       |   |  - OverlayFS      |   |  - veth pair          |
|  - Mounts nosuid/...  |   |  - Read-Only root |   |  - Bridge mc0 + NAT   |
|  - STRIDE Threat Model|   |  - Pseudo-FS (/p..|   |  - Default Gateway    |
+-----------------------+   +-------------------+   +-----------------------+
        |                             |                         |
        +-----------------------------+-------------------------+
                                      |
                                      v
                        +---------------------------+
                        |      CONTROL DE RECURSOS  |
                        |      - Cgroups v2 Unif.   |
                        |      - memory.max         |
                        |      - cpu.max (CFS)      |
                        |      - pids.max           |
                        +---------------------------+
```

### Cronología de los 20 Días

- **Días 01 a 05 (Semana 1 - Fundamentos):**
  - Análisis de `fork`, `execve` y `clone`.
  - Exploración de los 7+1 namespaces.
  - UTS Namespace (hostname) y PID Namespace (PID 1).
  - Mount Namespace y `/proc` privado.
  - CI Workflow, linters (`golangci-lint`) y tests unitarios.

- **Días 06 a 10 (Semana 2 - Aislamiento de Filesystem):**
  - Aislamiento de raíz con `chroot`.
  - Reemplazo seguro de raíz con `pivot_root`.
  - Montaje de pseudo-sistemas de archivos (`/proc`, `/sys`, `/dev`, `/dev/pts`).
  - Pruebas de integración privilegiadas en GitHub Actions.
  - Hardening de filesystem (rootfs read-only opcional y rollback seguro).

- **Días 11 a 15 (Semana 3 - Recursos, Seguridad y Red):**
  - Cgroups v2 (estructura jerárquica y OOM killer).
  - Control de recursos de memoria, CPU y procesos concurrentes desde Go.
  - Reducción del Bounding Set de Linux Capabilities y `PR_SET_NO_NEW_PRIVS`.
  - Network Namespace, interfaces virtuales `veth` y asignación de IPs.
  - Virtual Bridge `mc0`, enrutamiento IPv4 y NAT MASQUERADE con `iptables`.

- **Días 16 a 20 (Semana 4 - CLI, Inspección, Capas y Release):**
  - Diseño modular de la CLI y persistencia de estado daemonless (`/tmp/mc/containers/`).
  - Inspección y debug con `setns(2)` y `nsenter` (`mc exec`).
  - Capas de sistema de archivos con OverlayFS (`--overlay`).
  - Modelo formal de amenazas STRIDE y escáner de seguridad `gosec`.
  - Release oficial v0.1.0 con distribución multiarquitectura (`amd64`/`arm64`).

---

## Empaquetado y Verificación de la Release

Los binarios generados se compilan estáticamente y se empaquetan en archivos tar comprimidos con gzip:

```bash
# Verificar la suma criptográfica SHA-256
sha256sum -c checksums.txt
```

---

## Checklist de Cierre

- [x] Actualizado `mc version` para reportar versión formal `v0.1.0` y metadatos de entorno.
- [x] Configurado job de GitHub Actions para release multiarquitectura y generación de checksums.
- [x] Finalizada la documentación del proyecto y la matriz arquitectónica.
- [x] 100% de los 20 días completados según las pautas de genealogía lineal y validación por PR.
