# ZeroDayContainer (`mc`)

[![CI](https://github.com/GuamanJordan/ZeroDayContainer/actions/workflows/ci.yml/badge.svg)](https://github.com/GuamanJordan/ZeroDayContainer/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/GuamanJordan/ZeroDayContainer?color=blue&label=release)](https://github.com/GuamanJordan/ZeroDayContainer/releases)
![Go Version](https://img.shields.io/badge/go-1.22%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Security](https://img.shields.io/badge/security-gosec%20passed-brightgreen)

**ZeroDayContainer** (`mc`) es un runtime de contenedores en Linux modular y de alto rendimiento escrito en Go desde cero. Implementa aislamiento completo de procesos utilizando primitivas nativas del kernel Linux (**Namespaces**, **Cgroups v2**, **Linux Capabilities**, **PivotRoot**, **OverlayFS** y **Red Virtual**), sin depender de Docker, containerd ni `runc`.

---

## 🚀 Hoja de Ruta y Estado de Implementación (20 Días)

El proyecto ha completado de forma rigurosa la totalidad de los 20 hitos planificados con una genealogía lineal y verificación en CI:

- [x] **Día 01:** Análisis de llamadas de procesos (`fork`, `execve`, `clone`).
- [x] **Día 02:** Exploración práctica de los 7+1 Namespaces de Linux.
- [x] **Día 03:** Aislamiento UTS (hostname) y PID (PID 1 dentro del contenedor).
- [x] **Día 04:** Aislamiento Mount (`CLONE_NEWNS`) y pseudo-filesystem `/proc` privado.
- [x] **Día 05:** Flujo de CI automatizado, linters (`golangci-lint`) y pruebas unitarias.
- [x] **Día 06:** Aislamiento básico del sistema de archivos con `chroot`.
- [x] **Día 07:** Reemplazo atómico de la raíz del sistema de archivos con `pivot_root`.
- [x] **Día 08:** Montaje seguro de pseudo-filesystems esenciales (`/proc`, `/sys`, `/dev`, `/dev/pts`).
- [x] **Día 09:** Pruebas de integración privilegiadas en entornos virtualizados de CI.
- [x] **Día 10:** Hardening de filesystem (rootfs read-only opcional, desmontajes seguros y rollback).
- [x] **Día 11:** Arquitectura y jerarquía unificada de Cgroups v2.
- [x] **Día 12:** Control de recursos (memoria, CPU CFS quotas, límite de PIDs) desde Go.
- [x] **Día 13:** Reducción del Bounding Set de Linux Capabilities y activación de `PR_SET_NO_NEW_PRIVS`.
- [x] **Día 14:** Network namespace (`CLONE_NEWNET`), creación de interfaces virtuales `veth` y asignación de IPs.
- [x] **Día 15:** Conexión de bridge virtual `mc0`, reenvío `ip_forward` y NAT MASQUERADE con `iptables`.
- [x] **Día 16:** Diseño modular de CLI daemonless y seguimiento de estado en `/tmp/mc/containers/` (`mc list`, `mc version`).
- [x] **Día 17:** Inspección interactiva y ejecución en contenedores vivos mediante `setns(2)` y `nsenter` (`mc exec`).
- [x] **Día 18:** Sistemas de archivos en capas con OverlayFS (`lowerdir`, `upperdir`, `workdir`, `merged`) y bandera `--overlay`.
- [x] **Día 19:** Modelo formal de amenazas ([STRIDE Threat Model](docs/THREAT_MODEL.md)) y auditoría estática con Gosec.
- [x] **Día 20:** Release v0.1.0, compilación cruzada multiarquitectura (`linux/amd64`, `linux/arm64`) y verificación criptográfica (SHA-256).
- [x] **Día 21:** Montaje de volúmenes y Bind Mounts persistentes/compartidos (`-v, --volume`).

---

## 🛠️ Uso de la CLI (`mc`)

### Requisitos
- Linux con kernel ≥ 5.4 y privilegios de `root` (o WSL2 con soporte de Cgroups v2).
- Go 1.22+ (solo para compilación local).

### Compilación Local

```bash
go build -o mc ./cmd/mc
```

### Ejecución de Contenedores (`mc run`)

```bash
# 1. Contenedor básico aislado con pivot_root
sudo ./mc run /tmp/alpine-rootfs /bin/sh

# 2. Con capa de unión efímera OverlayFS (no altera la imagen base)
sudo ./mc run --overlay /tmp/alpine-rootfs /bin/sh

# 3. Con volúmenes bind mount del host (-v host:container[:ro])
sudo ./mc run -v /home/user/app:/app /tmp/alpine-rootfs /bin/sh

# 4. Modo de máxima seguridad con rootfs de solo lectura
sudo ./mc run --read-only /tmp/alpine-rootfs /bin/sh

# 5. Con límites de recursos garantizados por Cgroups v2
sudo ./mc run --memory=128m --cpus=1.0 --pids=50 /tmp/alpine-rootfs /bin/sh

# 6. Con pila de red aislada y veth pair
sudo ./mc run --net /tmp/alpine-rootfs /bin/sh

# 7. Con salida completa a Internet vía bridge mc0 y NAT
sudo ./mc run --nat /tmp/alpine-rootfs /bin/sh

# 8. Combinación completa para entorno de producción
sudo ./mc run --overlay --nat -v /home/user/app:/app --memory=256m --cpus=1.0 --pids=100 /tmp/alpine-rootfs /bin/sh
```

### Ejecutar Comandos en Contenedores Activos (`mc exec`)

```bash
# Conectarse a un contenedor por su ID asignado
sudo ./mc exec mc-123456 /bin/sh

# Ejecutar un comando puntual por PID directo del host
sudo ./mc exec 14209 /bin/ps aux
```

### Monitoreo y Estado (`mc list` o `mc ps`)

```bash
./mc list
```

Salida de ejemplo:
```text
CONTAINER ID    PID      STATUS     COMMAND              CREATED
mc-83921        14209    running    /bin/sh              2026-09-08 11:30:15
```

### Versión del Runtime (`mc version`)

```bash
./mc version
```

---

## 🔒 Arquitectura de Seguridad

Consulte el documento exhaustivo [Modelo de Amenazas y Seguridad](docs/THREAT_MODEL.md) para conocer los vectores analizados bajo el modelo STRIDE y la matriz comparativa de aislamiento contra Docker (`runc`), `gVisor` y `Kata Containers`.

---

## 🧪 Pruebas y Calidad de Código

### Pruebas Unitarias, Linters y Escaneo de Seguridad
```bash
go vet ./...
go test ./... -v -race -cover
gosec -exclude=G204,G304 ./...
```

### Pruebas de Integración (Requiere privilegios de root)
```bash
sudo go test ./test/integration/... -tags=integration -v
```

---

## 📄 Licencia

Este proyecto se distribuye bajo los términos de la Licencia MIT.
