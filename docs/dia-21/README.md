# Día 21: Montaje de Volúmenes y Bind Mounts (`-v`)

## Objetivo

Implementar el soporte para el montaje de volúmenes de almacenamiento mediante **Bind Mounts** en Linux, permitiendo compartir directorios o archivos individuales entre el sistema de archivos del host y el sistema de archivos raíz (`rootfs`) del contenedor en tiempo real, con soporte para modos de lectura/escritura (`rw`) y solo lectura (`ro`).

Al finalizar este día debes poder explicar:

- Cómo opera la llamada al sistema `mount(2)` con el flag `MS_BIND` y `MS_REC`.
- Cómo transformar un bind mount a modo de solo lectura atómicamente mediante un remount (`MS_BIND | MS_REMOUNT | MS_RDONLY`).
- La diferencia fundamental entre el almacenamiento efímero provisto por OverlayFS y el almacenamiento persistente/compartido mediante volúmenes.
- El uso de la bandera `-v` o `--volume <host>:<contenedor>[:ro|rw]` en la CLI `mc`.

---

## Anatomía de un Bind Mount en Linux

Un **Bind Mount** permite tomar un subárbol existente del sistema de archivos y hacerlo visible en un segundo punto de montaje, independientemente del tipo de sistema de archivos físico subyacente:

```text
Host Filesystem (/home/user/app)  ----------------+
                                                  |  syscall.Mount(..., MS_BIND)
                                                  v
Contenedor Rootfs (/app) <========================+
```

### Remount Read-Only (`:ro`)

El kernel Linux no permite crear directamente un bind mount en modo de solo lectura en una sola operación `mount`. Para lograrlo de forma segura y atómica, se realiza en dos pasos:
1. **Paso 1 (Creación):** Se realiza el enlace inicial con `MS_BIND | MS_REC`.
2. **Paso 2 (Endurecimiento):** Se aplica un remount sobre el punto de montaje destino con `MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY`.

---

## Sintaxis y Ejemplos de Uso en `mc`

### 1. Compartir código en tiempo de desarrollo (Lectura y Escritura)
```bash
sudo mc run -v /home/jordan/proyecto:/workspace /tmp/alpine-rootfs /bin/sh
```
*Cualquier archivo creado o modificado en `/workspace` dentro del contenedor se refleja instantáneamente en el host y sobrevive a la destrucción del contenedor.*

### 2. Inyectar configuración protegida (Solo Lectura `:ro`)
```bash
sudo mc run -v /etc/hosts:/etc/hosts:ro /tmp/alpine-rootfs /bin/sh
```
*El contenedor puede leer el archivo, pero cualquier intento de escritura resultará en un error de `Read-only file system`.*

---

## Checklist de Cierre

- [x] Implementado el paquete `internal/rootfs/volumes_linux.go` con `ParseVolumeSpec`, `MountVolumes` y `UnmountVolumes`.
- [x] Añadidos stubs multiplataforma en `internal/rootfs/volumes_others.go`.
- [x] Pruebas unitarias de parsing y validación en `internal/rootfs/volumes_test.go`.
- [x] Integración de la bandera `-v` / `--volume` en `cmd/mc/main.go` y `internal/namespaces/run_linux.go`.
- [x] Desmontaje automático (`MNT_DETACH`) en orden inverso al salir del contenedor.
