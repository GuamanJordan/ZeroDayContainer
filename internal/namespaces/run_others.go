//go:build !linux

package namespaces

import (
	"fmt"
	"syscall"
)

// GetBasicSysProcAttr en sistemas no-Linux retorna una estructura vacía.
func GetBasicSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

// RunBasic en sistemas no-Linux retorna un error indicando que se requiere Linux.
func RunBasic(cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar namespaces (CLONE_NEWUTS, CLONE_NEWPID)")
}

// RunChroot en sistemas no-Linux retorna un error de incompatibilidad.
func RunChroot(newRoot string, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar chroot")
}

// ChildInit en sistemas no-Linux retorna un error indicando que se requiere Linux.
func ChildInit(cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar child-init")
}

// ChildInitChroot en sistemas no-Linux retorna un error indicando que se requiere Linux.
func ChildInitChroot(newRoot string, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar child-init-chroot")
}
