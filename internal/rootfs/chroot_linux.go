//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"syscall"
)

// ApplyChroot cambia la raíz del sistema de archivos al directorio newRoot
// y establece el directorio de trabajo actual a "/" mediante syscall.Chdir.
func ApplyChroot(newRoot string) error {
	if newRoot == "" {
		return fmt.Errorf("la ruta de rootfs no puede estar vacía")
	}

	info, err := os.Stat(newRoot)
	if err != nil {
		return fmt.Errorf("error al acceder a la ruta de rootfs '%s': %w", newRoot, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("la ruta de rootfs '%s' no es un directorio", newRoot)
	}

	if err := syscall.Chroot(newRoot); err != nil {
		return fmt.Errorf("syscall.Chroot('%s') falló: %w", newRoot, err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("syscall.Chdir('/') tras chroot falló: %w", err)
	}

	return nil
}
