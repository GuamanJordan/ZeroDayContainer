package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// RunBasic lanza `cmdPath args...` en un proceso hijo aislado
// con su propio UTS namespace (hostname) y PID namespace.
func RunBasic(cmdPath string, args []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"child-init", cmdPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	return cmd.Run()
}

// ChildInit corre YA DENTRO del namespace nuevo (PID 1 del árbol nuevo).
// Se invoca a sí mismo vía /proc/self/exe, ver main.go -> caso "child-init".
func ChildInit(cmdPath string, args []string) error {
	if err := syscall.Sethostname([]byte("zerodaycontainer")); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
