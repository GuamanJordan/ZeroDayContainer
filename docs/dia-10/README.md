# Día 10: Hardening de Filesystem — Rootfs Read-Only y Desmontajes Seguros

## Objetivo

Robustecer la capa del sistema de archivos del contenedor implementando técnicas esenciales de **hardening** y **limpieza defensiva**:
1. Permitir ejecutar el rootfs del contenedor en modo de **solo lectura** (`MS_RDONLY`) para evitar manipulaciones o persistencia no autorizada.
2. Implementar **desmontajes seguros y mecanismos de rollback** ante fallos de montaje usando `defer`, evitando filtraciones de puntos de montaje huérfanos en el host o en el espacio de nombres.
3. Documentar el ciclo de vida completo de preparación, `pivot_root`, montajes y desmontajes con diagramas visuales.

Al finalizar este día debes poder explicar:

- Por qué los runtimes modernos de contenedores (como Docker, containerd o Podman) admiten banderas como `--read-only`.
- Cómo funciona `syscall.Mount("", target, "", MS_REMOUNT|MS_BIND|MS_RDONLY, "")` a nivel del kernel VFS.
- Cómo gestionar el ciclo de vida de fallos mediante funciones de rollback diferidas (`defer`).
- La diferencia entre un desmontaje estándar y `MNT_DETACH` (lazy unmount).

---

## Hardening de Filesystem: Rootfs de Solo Lectura (`MS_RDONLY`)

### ¿Por qué montar el rootfs como Read-Only?
Por defecto, si un contenedor se ejecuta con permisos de escritura sobre su raíz, cualquier proceso comprometido dentro del contenedor puede alterar archivos de configuración (`/etc/passwd`, `/etc/shadow`), binarios del sistema (`/bin/sh`), o descargar malware ejecutable.

Al activar `MS_RDONLY`:
- La raíz `/` queda sellada contra cualquier modificación en tiempo de ejecución.
- Si una aplicación requiere almacenamiento efímero (como `/tmp`), se le monta un `tmpfs` explícito independiente.
- Se implementa el principio de **inmutabilidad de infraestructura**.

### Mecánica de Kernel: Remount con `MS_REMOUNT`
Para hacer un rootfs de solo lectura cuando se usa `pivot_root`:
1. Primero se realiza el `pivot_root` normalmente (el nuevo rootfs debe ser accesible para lectura y configuración).
2. Luego se invoca una operación de remontaje sobre `/`:
```go
func RemountReadOnly(target string) error {
	return syscall.Mount("", target, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_RDONLY, "")
}
```

---

## Desmontajes Seguros y Rollback ante Errores

Si un contenedor falla al montar uno de sus pseudo-sistemas de archivos (por ejemplo, si `/proc` se monta correctamente pero `/sys` falla), el espacio de nombres de montaje no debe dejar recursos inconsistentes.

### Patrón de Rollback Implementado
En `internal/rootfs/mounts_linux.go`:
```go
func MountEssentialFilesystems() (err error) {
	var mounted []string
	defer func() {
		if err != nil {
			RollbackMounts(mounted)
		}
	}()

	for _, m := range GetEssentialMounts() {
		// Montar y registrar en slice 'mounted'
		...
		mounted = append(mounted, m.Target)
	}
	return nil
}
```

Al salir del contenedor o ante un error fatal, `UnmountEssentialFilesystems()` y `RollbackMounts()` desmontan los puntos en **orden inverso**, asegurando que los sistemas montados encima de otros se liberen limpiamente.

---

## Diagrama del Ciclo de Vida Completo de `pivot_root` y Hardening

### Diagrama Mermaid

```mermaid
flowchart TD
    A[mc run / mc run --read-only] --> B[Lanzar proceso hijo con CLONE_NEWUTS, CLONE_NEWPID, CLONE_NEWNS]
    B --> C[child-init-pivot: Sethostname 'zerodaycontainer']
    C --> D[syscall.Mount MS_PRIVATE en /]
    D --> E[Bind-mount de newRoot sobre sí mismo]
    E --> F[Crear directorio temporal .old_root]
    F --> G[syscall.PivotRoot: newRoot -> /, .old_root -> raíz antigua]
    G --> H[syscall.Chdir /]
    H --> I[syscall.Unmount /.old_root con MNT_DETACH]
    I --> J[os.Remove /.old_root]
    J --> K{¿Opción ReadOnly activa?}
    K -- Sí --> L[Remontar / con MS_REMOUNT | MS_BIND | MS_RDONLY]
    K -- No --> M[Mantener permisos RW en /]
    L --> N[Montar sistemas esenciales: /proc, /sys, /dev, /dev/pts]
    M --> N
    N -- Error en montajes --> O[Rollback en orden inverso: desmontar recursos previos]
    N -- Éxito --> P[Ejecutar comando del usuario en PID 1]
    P --> Q[defer: UnmountEssentialFilesystems al finalizar]
```

### Flujo ASCII

```text
Host FS                          Nuevo Rootfs                          Contenedor
=======                          ============                          ==========
   |                                  |                                     |
   +--- Bind-mount newRoot ---------->|                                     |
   |                                  +--- Crear .old_root                  |
   |                                  |                                     |
   +<--- Mover raíz vieja a .old_root-+                                     |
   |                                  +--- Convertir newRoot en / --------->|
   |                                                                        +--- Chdir("/")
   +<--- Unmount MNT_DETACH /.old_root y Remove                             |
                                                                            +--- [Opcional] Remount / (MS_RDONLY)
                                                                            |
                                                                            +--- Montajes (/proc, /sys, /dev, /dev/pts)
                                                                            |    (con Rollback automático en fallo)
                                                                            |
                                                                            +--- Exec comando del usuario
                                                                            |
                                                                            +--- Defer Cleanup & Unmounts
```

---

## Uso desde la CLI (`mc run`)

### Ejecutar contenedor interactivo normal (lectura/escritura):
```bash
sudo ./mc run /tmp/alpine-rootfs /bin/sh
```

### Ejecutar contenedor con rootfs de solo lectura (`--read-only` o `-ro`):
```bash
sudo ./mc run --read-only /tmp/alpine-rootfs /bin/sh
```

Dentro del contenedor en modo solo lectura, cualquier intento de escritura en el sistema de archivos raíz será bloqueado por el kernel:
```sh
/ # touch /test.txt
touch: /test.txt: Read-only file system
```

---

## Checklist de Cierre

- [x] Agregada estructura `PivotOptions` y función `ApplyPivotRootWithOptions` en `internal/rootfs`.
- [x] Implementada la función `RemountReadOnly` para remontar con `MS_REMOUNT|MS_BIND|MS_RDONLY`.
- [x] Implementado rollback automático en `MountEssentialFilesystems` y función `UnmountEssentialFilesystems`.
- [x] Añadida bandera `--read-only` (`-ro`) en `cmd/mc/main.go` y propagación en `internal/namespaces`.
- [x] Pruebas unitarias de validación y rollback en `internal/rootfs`.
- [x] Prueba de integración con `--read-only` en `test/integration/integration_test.go`.
