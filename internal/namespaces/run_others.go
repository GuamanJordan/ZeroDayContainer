//go:build !linux

package namespaces

import (
	"fmt"
	"syscall"

	"github.com/GuamanJordan/ZeroDayContainer/internal/cgroups"
	"github.com/GuamanJordan/ZeroDayContainer/internal/network"
	"github.com/GuamanJordan/ZeroDayContainer/internal/rootfs"
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

// ContainerOpts define las opciones de aislamiento.
type ContainerOpts struct {
	ReadOnly      bool
	EnableNet     bool
	EnableNAT     bool
	EnableOverlay bool
	Volumes       []rootfs.VolumeMount
	PortMappings  []network.PortMapping
	Cgroups       cgroups.Config
}

// RunPivotRoot en sistemas no-Linux retorna un error de incompatibilidad.
func RunPivotRoot(newRoot string, readOnly bool, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar pivot_root")
}

// RunPivotRootWithCgroups en sistemas no-Linux retorna un error de incompatibilidad.
func RunPivotRootWithCgroups(newRoot string, readOnly bool, cgCfg cgroups.Config, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar cgroups y pivot_root")
}

// RunPivotRootWithOptions en sistemas no-Linux retorna un error de incompatibilidad.
func RunPivotRootWithOptions(newRoot string, opts ContainerOpts, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar pivot_root con opciones")
}

// ChildInit en sistemas no-Linux retorna un error indicando que se requiere Linux.
func ChildInit(cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar child-init")
}

// ChildInitChroot en sistemas no-Linux retorna un error indicando que se requiere Linux.
func ChildInitChroot(newRoot string, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar child-init-chroot")
}

// ChildInitPivotRoot en sistemas no-Linux retorna un error indicando que se requiere Linux.
func ChildInitPivotRoot(newRoot string, readOnly bool, cmdPath string, args []string) error {
	return fmt.Errorf("ZeroDayContainer requiere Linux para ejecutar child-init-pivot")
}
