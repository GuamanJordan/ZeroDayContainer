//go:build !linux

package rootfs

import (
	"fmt"
)

// PivotOptions define opciones de configuración adicionales para pivot_root.
type PivotOptions struct {
	ReadOnly bool
}

// ApplyPivotRoot en plataformas no-Linux retorna un error indicando incompatibilidad.
func ApplyPivotRoot(newRoot string) error {
	return fmt.Errorf("ApplyPivotRoot requiere Linux para ejecutarse")
}

// ApplyPivotRootWithOptions en plataformas no-Linux retorna un error indicando incompatibilidad.
func ApplyPivotRootWithOptions(newRoot string, opts PivotOptions) error {
	return fmt.Errorf("ApplyPivotRootWithOptions requiere Linux para ejecutarse")
}

// RemountReadOnly en plataformas no-Linux retorna un error indicando incompatibilidad.
func RemountReadOnly(target string) error {
	return fmt.Errorf("RemountReadOnly requiere Linux para ejecutarse")
}
