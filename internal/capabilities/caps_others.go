//go:build !linux

package capabilities

import "fmt"

// GetDefaultDangerousCapabilities retorna la lista de capabilities en plataformas no-Linux.
func GetDefaultDangerousCapabilities() []int {
	return nil
}

// EnableNoNewPrivs en plataformas no-Linux retorna un error de incompatibilidad.
func EnableNoNewPrivs() error {
	return fmt.Errorf("capabilities y prctl solo están soportados en Linux")
}

// DropBoundingCapability en plataformas no-Linux retorna un error de incompatibilidad.
func DropBoundingCapability(cap int) error {
	return fmt.Errorf("capabilities y prctl solo están soportados en Linux")
}

// DropDangerousCapabilities en plataformas no-Linux retorna un error de incompatibilidad.
func DropDangerousCapabilities() error {
	return fmt.Errorf("capabilities y prctl solo están soportados en Linux")
}

// RestrictPrivileges en plataformas no-Linux retorna un error de incompatibilidad.
func RestrictPrivileges() error {
	return fmt.Errorf("capabilities y prctl solo están soportados en Linux")
}
