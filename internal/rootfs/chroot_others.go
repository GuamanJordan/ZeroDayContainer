//go:build !linux

package rootfs

import (
	"fmt"
)

// ApplyChroot en plataformas no-Linux retorna un error indicando incompatibilidad.
func ApplyChroot(newRoot string) error {
	return fmt.Errorf("ApplyChroot requiere Linux para ejecutarse")
}
