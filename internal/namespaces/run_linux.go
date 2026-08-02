//go:build linux

package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// GetBasicSysProcAttr retorna la configuración SysProcAttr con las flags
// CLONE_NEWUTS y CLONE_NEWPID activadas para aislamiento básico.
func GetBasicSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}
}

// RunBasic lanza `cmdPath args...` en un proceso hijo aislado
// con su propio UTS namespace (hostname) y PID namespace.
func RunBasic(cmdPath string, args []string) error {
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar")
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"child-init", cmdPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = GetBasicSysProcAttr()

	return cmd.Run()
}

// ChildInit se ejecuta dentro del nuevo namespace (PID 1 del nuevo árbol).
// Configura el hostname y ejecuta el binario final mediante exec.
func ChildInit(cmdPath string, args []string) error {
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar dentro del contenedor")
	}

	if err := syscall.Sethostname([]byte("zerodaycontainer")); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
