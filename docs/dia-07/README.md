# Día 7: pivot_root — El Estándar de Aislamiento de Raíz en Contenedores

## Objetivo

Comprender el funcionamiento de `pivot_root()`, aprender la secuencia de montajes necesarios para sustituir el sistema de archivos raíz del host de forma segura, y entender por qué los runtimes como OCI (`runc`, Docker, Containerd) utilizan `pivot_root` en lugar de `chroot`.

Al finalizar este día debes poder explicar:

- Las diferencias técnicas clave entre `chroot` y `pivot_root`.
- Por qué `pivot_root` requiere que `new_root` sea un *mount point*.
- Los 7 pasos exactos para cambiar de raíz de forma segura sin dejar rastro de la vieja raíz del host.
- Cómo usar el subcomando `mc run <rootfs> <comando>`.

---

## Comparativa: `chroot` vs `pivot_root`

| Característica | `chroot` | `pivot_root` |
| --- | --- | --- |
| **Mecanismo** | Cambia el puntero del directorio raíz aparente del proceso. | Intercambia el punto de montaje del sistema de archivos raíz entero en el mount namespace. |
| **Aislamiento** | Débil (los montajes del host siguen en la tabla de montajes). | Fuerte (desvincula la raíz del host completamente con `umount`). |
| **Seguridad** | Posibilidad de escapes con `fchdir` / `..`. | Imposible escapar a la raíz del host porque la vieja raíz es desmontada. |
| **Requisito** | Funciona en cualquier directorio. | Requiere que `new_root` sea un *mount point* y estar en un Mount namespace propio. |

---

## Los 7 Pasos de `pivot_root`

1. **Cambiar propagación de montajes a `MS_PRIVATE`:** Evita que los cambios de montaje se propaguen al host.
2. **Hacer bind-mount de `newRoot` sobre sí mismo:** Garantiza que `newRoot` sea formalmente un punto de montaje.
3. **Crear directorio temporal `.old_root`:** Ubicado dentro de `newRoot` para alojar temporalmente la vieja raíz del host.
4. **Ejecutar `syscall.PivotRoot(newRoot, oldRoot)`:** Intercambia las raíces en el kernel.
5. **Cambiar directorio de trabajo a `/`:** `syscall.Chdir("/")`.
6. **Desmontar la vieja raíz:** `syscall.Unmount("/.old_root", syscall.MNT_DETACH)` (*lazy unmount*).
7. **Eliminar el directorio `.old_root`:** `os.Remove("/.old_root")`.

---

## Implementación en Go (`internal/rootfs/pivot_root_linux.go`)

```go
func ApplyPivotRoot(newRoot string) error {
	_ = syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	_ = syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, "")

	oldRootDir := filepath.Join(newRoot, ".old_root")
	_ = os.MkdirAll(oldRootDir, 0700)

	if err := syscall.PivotRoot(newRoot, oldRootDir); err != nil {
		return err
	}
	_ = syscall.Chdir("/")

	_ = syscall.Unmount("/.old_root", syscall.MNT_DETACH)
	_ = os.Remove("/.old_root")

	return nil
}
```

---

## Uso de la CLI `mc`

```bash
sudo ./mc run /tmp/alpine-rootfs /bin/sh
```

Dentro del contenedor:
- `ls /` mostrará única y exclusivamente el contenido del rootfs Alpine.
- `cat /proc/mounts` verificará que `/` apunta al nuevo punto de montaje y la raíz antigua del host ha sido desmontada totalmente.

---

## Checklist de Cierre

- [x] Implementado el paquete `internal/rootfs` con `ApplyPivotRoot`.
- [x] Creado el subcomando principal `mc run <rootfs> <cmd>`.
- [x] Implementado el bind-mount automático de `newRoot` sobre sí mismo.
- [x] Verificado el desmontaje seguro (`MNT_DETACH`) de `.old_root`.
- [x] Pruebas unitarias integradas en `internal/rootfs/pivot_root_test.go`.
