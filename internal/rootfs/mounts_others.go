//go:build !linux

package rootfs

import (
	"fmt"
)

// MountSpec define la estructura para especificar un montaje pseudo o de sistema de archivos.
type MountSpec struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}

// GetEssentialMounts en sistemas no-Linux retorna una lista vacía.
func GetEssentialMounts() []MountSpec {
	return nil
}

// MountEssentialFilesystems en sistemas no-Linux retorna un error de incompatibilidad.
func MountEssentialFilesystems() error {
	return fmt.Errorf("MountEssentialFilesystems requiere Linux para ejecutarse")
}
