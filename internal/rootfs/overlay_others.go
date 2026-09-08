//go:build !linux

package rootfs

import "fmt"

// OverlayConfig define las rutas requeridas para inicializar un sistema de archivos OverlayFS.
type OverlayConfig struct {
	LowerDir  string
	UpperDir  string
	WorkDir   string
	MergedDir string
}

// SetupOverlayDirectories en sistemas no-Linux retorna error de incompatibilidad.
func SetupOverlayDirectories(cfg OverlayConfig) error {
	return fmt.Errorf("OverlayFS solo está soportado en Linux")
}

// MountOverlay en sistemas no-Linux retorna error de incompatibilidad.
func MountOverlay(cfg OverlayConfig) error {
	return fmt.Errorf("OverlayFS solo está soportado en Linux")
}

// UnmountOverlay en sistemas no-Linux retorna error de incompatibilidad.
func UnmountOverlay(mergedDir string) error {
	return fmt.Errorf("OverlayFS solo está soportado en Linux")
}
