//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// PivotOptions define opciones de configuración adicionales para el montaje del nuevo rootfs.
type PivotOptions struct {
	ReadOnly bool
}

// ApplyPivotRoot cambia el directorio raíz del proceso actual utilizando syscall.PivotRoot con opciones por defecto.
func ApplyPivotRoot(newRoot string) error {
	return ApplyPivotRootWithOptions(newRoot, PivotOptions{ReadOnly: false})
}

// ApplyPivotRootWithOptions cambia el directorio raíz del proceso actual utilizando syscall.PivotRoot
// y permite aplicar opciones de hardening como montar la nueva raíz en modo solo lectura (MS_RDONLY).
// Realiza los pasos requeridos garantizando la limpieza segura de recursos temporales y montajes ante fallos:
// 1. Establece la propagación de montajes a privado (MS_PRIVATE).
// 2. Realiza un bind-mount de newRoot sobre sí mismo para garantizar que sea un punto de montaje.
// 3. Crea el directorio temporal .old_root dentro de newRoot.
// 4. Ejecuta syscall.PivotRoot(newRoot, oldRoot).
// 5. Cambia el directorio de trabajo a "/".
// 6. Desmonta la raíz antigua (.old_root) con MNT_DETACH y elimina el directorio temporal.
// 7. Si opts.ReadOnly es verdadero, remonta "/" en modo solo lectura con MS_RDONLY.
func ApplyPivotRootWithOptions(newRoot string, opts PivotOptions) (err error) {
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
	bindMounted := true
	defer func() {
		if err != nil && bindMounted {
			_ = syscall.Unmount(newRoot, syscall.MNT_DETACH)
		}
	}()

	// 3. Crear directorio temporal para la raíz antigua dentro del nuevo rootfs
	oldRootDir := filepath.Join(newRoot, ".old_root")
	if err := os.MkdirAll(oldRootDir, 0700); err != nil {
		return fmt.Errorf("mkdir .old_root en '%s': %w", oldRootDir, err)
	}
	defer func() {
		if err != nil && bindMounted {
			_ = os.Remove(oldRootDir)
		}
	}()

	// 4. Pivotar el sistema de archivos raíz
	if err := syscall.PivotRoot(newRoot, oldRootDir); err != nil {
		return fmt.Errorf("syscall.PivotRoot('%s', '%s') falló: %w", newRoot, oldRootDir, err)
	}
	bindMounted = false // Una vez ejecutado pivot_root exitosamente, newRoot pasa a ser "/"

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

	// 8. Aplicar hardening de rootfs de solo lectura si fue solicitado
	if opts.ReadOnly {
		if err := RemountReadOnly("/"); err != nil {
			return fmt.Errorf("remontar rootfs en solo lectura: %w", err)
		}
	}

	return nil
}

// RemountReadOnly remonta el punto de montaje destino con la flag MS_RDONLY para impedir escrituras.
func RemountReadOnly(target string) error {
	if target == "" {
		return fmt.Errorf("la ruta destino para remontar solo lectura no puede estar vacía")
	}
	if err := syscall.Mount("", target, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_RDONLY, ""); err != nil {
		return fmt.Errorf("syscall.Mount MS_REMOUNT|MS_RDONLY en '%s' falló: %w", target, err)
	}
	return nil
}
