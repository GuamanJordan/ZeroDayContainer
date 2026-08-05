//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// ApplyPivotRoot cambia el directorio raíz del proceso actual utilizando la syscall.PivotRoot.
// Realiza los pasos de preparación requeridos por el kernel:
// 1. Establece la propagación de montajes a privado (MS_PRIVATE).
// 2. Realiza un bind-mount de newRoot sobre sí mismo para garantizar que sea un punto de montaje.
// 3. Crea el directorio temporal .old_root dentro de newRoot.
// 4. Ejecuta syscall.PivotRoot(newRoot, oldRoot).
// 5. Cambia el directorio de trabajo a "/".
// 6. Desmonta la raíz antigua (.old_root) con MNT_DETACH y elimina el directorio temporal.
func ApplyPivotRoot(newRoot string) error {
	if newRoot == "" {
		return fmt.Errorf("la ruta del nuevo rootfs no puede estar vacía")
	}

	info, err := os.Stat(newRoot)
	if err != nil {
		return fmt.Errorf("error al acceder a la ruta de rootfs '%s': %w", newRoot, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("la ruta de rootfs '%s' no es un directorio", newRoot)
	}

	// 1. Asegurar propagación privada en el espacio de nombres de montaje
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("mount MS_PRIVATE en '/': %w", err)
	}

	// 2. Garantizar que newRoot sea un punto de montaje (bind-mount sobre sí mismo)
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount MS_BIND de newRoot '%s': %w", newRoot, err)
	}

	// 3. Crear directorio temporal para la raíz antigua dentro del nuevo rootfs
	oldRootDir := filepath.Join(newRoot, ".old_root")
	if err := os.MkdirAll(oldRootDir, 0700); err != nil {
		return fmt.Errorf("mkdir .old_root en '%s': %w", oldRootDir, err)
	}

	// 4. Pivotar el sistema de archivos raíz
	if err := syscall.PivotRoot(newRoot, oldRootDir); err != nil {
		return fmt.Errorf("syscall.PivotRoot('%s', '%s') falló: %w", newRoot, oldRootDir, err)
	}

	// 5. Cambiar el directorio de trabajo actual a "/"
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("syscall.Chdir('/') tras pivot_root falló: %w", err)
	}

	// 6. Desmontar la raíz antigua de forma perezosa (lazy unmount)
	oldRootPath := "/.old_root"
	if err := syscall.Unmount(oldRootPath, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("syscall.Unmount('%s') falló: %w", oldRootPath, err)
	}

	// 7. Eliminar el punto de montaje temporal
	if err := os.Remove(oldRootPath); err != nil {
		return fmt.Errorf("os.Remove('%s') falló: %w", oldRootPath, err)
	}

	return nil
}
