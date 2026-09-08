//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"syscall"
)

// OverlayConfig define las rutas requeridas para inicializar un sistema de archivos OverlayFS.
type OverlayConfig struct {
	LowerDir  string
	UpperDir  string
	WorkDir   string
	MergedDir string
}

// SetupOverlayDirectories crea los directorios requeridos para las capas de OverlayFS.
func SetupOverlayDirectories(cfg OverlayConfig) error {
	if cfg.LowerDir == "" {
		return fmt.Errorf("se debe especificar lowerdir")
	}
	if cfg.UpperDir == "" || cfg.WorkDir == "" || cfg.MergedDir == "" {
		return fmt.Errorf("upperdir, workdir y merged deben ser especificados")
	}

	for _, dir := range []string{cfg.UpperDir, cfg.WorkDir, cfg.MergedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error al crear directorio overlay %s: %w", dir, err)
		}
	}
	return nil
}

// MountOverlay monta un sistema de archivos OverlayFS unificando la capa base (lowerdir)
// con la capa de lectura/escritura (upperdir) usando workdir como área de staging.
func MountOverlay(cfg OverlayConfig) error {
	if err := SetupOverlayDirectories(cfg); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.LowerDir); os.IsNotExist(err) {
		return fmt.Errorf("lowerdir %s no existe: %w", cfg.LowerDir, err)
	}

	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", cfg.LowerDir, cfg.UpperDir, cfg.WorkDir)
	if err := syscall.Mount("overlay", cfg.MergedDir, "overlay", 0, data); err != nil {
		return fmt.Errorf("falló mount overlay en %s con data %q: %w", cfg.MergedDir, data, err)
	}

	return nil
}

// UnmountOverlay desmonta de forma segura el punto de unión OverlayFS (mergedDir).
func UnmountOverlay(mergedDir string) error {
	if mergedDir == "" {
		return fmt.Errorf("mergedDir no puede estar vacío")
	}
	if err := syscall.Unmount(mergedDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("falló desmontaje de overlay en %s: %w", mergedDir, err)
	}
	return nil
}
