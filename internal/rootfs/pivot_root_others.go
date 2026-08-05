//go:build !linux

package rootfs

import (
	"fmt"
)

// ApplyPivotRoot en plataformas no-Linux retorna un error indicando incompatibilidad.
func ApplyPivotRoot(newRoot string) error {
	return fmt.Errorf("ApplyPivotRoot requiere Linux para ejecutarse")
}
