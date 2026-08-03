# Día 6: chroot — La Abuela del Aislamiento de Filesystem

## Objetivo

Comprender el funcionamiento de `chroot()`, aprender a aislar el sistema de archivos raíz de un proceso a un directorio arbitrario (*rootfs*), y analizar por qué `chroot()` por sí solo no es seguro para contenedores modernos.

Al finalizar este día debes poder explicar:

- Qué hace la llamada al sistema `chroot()`.
- Por qué siempre debe ir acompañada de `chdir("/")`.
- Cómo usar el subcomando `mc run-chroot <rootfs> <comando>`.
- Las limitaciones y vulnerabilidades de seguridad de `chroot()` frente a `pivot_root()`.

---

## Modelo Mental: ¿Qué es `chroot`?

La llamada al sistema `chroot(const char *path)` (*change root*) cambia el directorio raíz percibido de un proceso (y de sus hijos futuros) a la ruta especificada.

Sin `chroot`, todo proceso ve `/` apuntando al sistema de archivos raíz real del host. Con `chroot("/home/user/alpine-rootfs")`, para ese proceso la ruta `/` apuntará a dicho directorio.

---

## Regla de Oro: `chroot` + `chdir("/")`

Invocar únicamente `syscall.Chroot(newRoot)` modifica el puntero de la nueva raíz de archivos, **pero no cambia el directorio de trabajo actual (CWD) del proceso**.

Si no ejecutas `syscall.Chdir("/")` inmediatamente después de `chroot`:
1. El proceso seguirá teniendo su directorio actual fuera de la nueva raíz.
2. Un proceso atacante podría navegar hacia arriba usando `..` o llamadas a `fchdir` sobre descriptores de archivos abiertos previos al chroot y **escapar del aislamiento**.

### Implementación en Go (`internal/rootfs/chroot_linux.go`)

```go
func ApplyChroot(newRoot string) error {
	if err := syscall.Chroot(newRoot); err != nil {
		return fmt.Errorf("syscall.Chroot: %w", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("syscall.Chdir: %w", err)
	}
	return nil
}
```

---

## Limitaciones de Seguridad de `chroot`

A diferencia de `pivot_root` (que se implementará en el Día 07), `chroot` no desvincula la vieja raíz del sistema de archivos del espacio de nombres de montaje.

1. **Escape con `fchdir` / `open` preexistente:** Si un proceso posee un descriptor de archivo abierto a un directorio fuera de la nueva raíz antes de ejecutar el chroot, puede usar `fchdir(fd)` para regresar al sistema de archivos del host.
2. **Dependencia de binarios en el rootfs:** Al cambiar la raíz a `newRoot`, el comando a ejecutar (ej. `/bin/sh` o `ls`) debe existir **dentro** del directorio `newRoot` junto con sus bibliotecas compartidas (`/lib`, `/lib64`), de lo contrario el `execve` fallará con *No such file or directory*.

---

## Uso de la CLI `mc`

### Requisito previo: Tener un rootfs mínimo (ej. Alpine Mini RootFS)

```bash
mkdir -p /tmp/alpine-rootfs
tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz -C /tmp/alpine-rootfs
```

### Ejecución aislada con `mc run-chroot`

```bash
sudo ./mc run-chroot /tmp/alpine-rootfs /bin/sh
```

Dentro del contenedor:
- `ls /` mostrará únicamente las carpetas del sistema de archivos Alpine (`bin`, `etc`, `lib`, `usr`, etc.).
- `hostname` mostrará `zerodaycontainer`.
- `ps aux` sólo mostrará los procesos del contenedor (gracias al Mount + PID + UTS namespace).

---

## Checklist de Cierre

- [x] Implementado el paquete `internal/rootfs` con `ApplyChroot`.
- [x] Creado el subcomando `mc run-chroot <rootfs> <cmd>`.
- [x] Verificada la llamada combinada `chroot(newRoot)` + `chdir("/")`.
- [x] Documentadas las limitaciones de seguridad de `chroot`.
- [x] Creados tests unitarios para la validación de rutas de `ApplyChroot`.
