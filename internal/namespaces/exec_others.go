//go:build !linux

package namespaces

import (
	"fmt"
)

// ExecInContainer en sistemas no-Linux retorna un error de incompatibilidad.
func ExecInContainer(targetPID int, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar setns/nsenter")
}
