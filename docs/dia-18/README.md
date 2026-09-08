# Día 18: Sistemas de Archivos en Capas: OverlayFS

## Objetivo

Comprender e implementar sistemas de archivos de unión (*union filesystems*) mediante **OverlayFS** en Linux, permitiendo que múltiples contenedores compartan una misma imagen base de solo lectura (`lowerdir`) mientras disponen de una capa efímera o persistente de lectura y escritura (`upperdir`), garantizando inmutabilidad de la base y optimización del almacenamiento.

Al finalizar este día debes poder explicar:

- Las cuatro entidades fundamentales de OverlayFS: `lowerdir`, `upperdir`, `workdir` y `merged`.
- El mecanismo de **Copy-Up** cuando un proceso intenta modificar un archivo existente en la capa inferior.
- El concepto y funcionamiento de los archivos especiales **Whiteout** (`character device 0/0`) para representar eliminaciones en la vista combinada.
- La integración de la bandera `--overlay` en `mc run` para ejecutar contenedores efímeros y seguros.

---

## Anatomía de OverlayFS

OverlayFS combina dos o más árboles de directorios en un único punto de montaje unificado:

```text
       +---------------------------------------------+
       |             merged (Vista Unificada)        |
       +---------------------------------------------+
                             ^
                             |
             +---------------+---------------+
             |                               |
     +---------------+               +---------------+
     |   upperdir    |               |   lowerdir    |
     |  (Lectura/    |               | (Solo Lectura,|
     |   Escritura)  |               |  Imagen Base) |
     +---------------+               +---------------+
             |
     +---------------+
     |    workdir    |  (Directorio interno de staging para
     |               |   operaciones atómicas del kernel)
     +---------------+
```

### Opciones de Montaje

La llamada al sistema `mount(2)` para OverlayFS requiere pasar las rutas en la cadena de datos `data`:

```c
mount("overlay", "/merged", "overlay", 0, "lowerdir=/base,upperdir=/diff,workdir=/work");
```

> **Requisito del Kernel:** `upperdir` y `workdir` deben pertenecer obligatoriamente al mismo sistema de archivos físico o lógico para garantizar la atomicidad de las operaciones `rename` y `copy-up`.

---

## Copy-Up y Whiteouts

1. **Lectura Directa:** Si un archivo existe sólo en `lowerdir`, el proceso lo lee directamente sin penalización de copia.
2. **Copy-Up en Escritura:** Si un archivo está en `lowerdir` y un proceso lo abre con modo `O_WRONLY` o `O_RDWR`, el kernel Linux copia íntegramente el archivo desde `lowerdir` a `upperdir` antes de permitir la escritura. Las modificaciones posteriores sólo afectan a la copia en `upperdir`.
3. **Eliminación con Whiteout:** Si se elimina un archivo originario de `lowerdir`, no se puede borrar de la capa base. En su lugar, el kernel crea un dispositivo de caracteres con major 0 y minor 0 (`0:0`) en `upperdir` con el mismo nombre. Cuando el VFS lee el directorio `merged`, la presencia del *whiteout* oculta el archivo subyacente.

---

## Uso con la CLI `mc`

Para ejecutar un contenedor con capa OverlayFS automática y efímera:

```bash
sudo mc run --overlay /tmp/alpine-rootfs /bin/sh
```

El runtime:
1. Crea un espacio de trabajo único en `/tmp/mc/overlay-<timestamp>/`.
2. Prepara los directorios `upper`, `work` y `merged`.
3. Monta el OverlayFS uniendo `/tmp/alpine-rootfs` (lower) con el `upper` recién creado.
4. Aplica `pivot_root` apuntando a `merged`.
5. Al salir el contenedor, desmonta limpiamente `merged` (`syscall.MNT_DETACH`) y elimina los datos efímeros.

---

## Checklist de Cierre

- [x] Implementado el paquete `internal/rootfs/overlay_linux.go` con `MountOverlay`, `UnmountOverlay` y `SetupOverlayDirectories`.
- [x] Añadidos stubs multiplataforma en `internal/rootfs/overlay_others.go`.
- [x] Pruebas unitarias de configuración y validaciones en `internal/rootfs/overlay_test.go`.
- [x] Integración de la bandera `--overlay` en `cmd/mc/main.go` y `internal/namespaces/run_linux.go`.
- [x] Limpieza automática (`defer`) de los montajes y directorios de staging al terminar el contenedor.
