//go:build linux

package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/GuamanJordan/ZeroDayContainer/internal/rootfs"
)

// GetBasicSysProcAttr retorna la configuración SysProcAttr con las flags
// CLONE_NEWUTS, CLONE_NEWPID y CLONE_NEWNS activadas para aislamiento de hostname, PIDs y montajes.
func GetBasicSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
}

// RunBasic lanza `cmdPath args...` en un proceso hijo aislado
// con UTS, PID y Mount namespaces.
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

// RunChroot lanza `cmdPath args...` cambiando la raíz del sistema de archivos al directorio rootfs usando chroot.
func RunChroot(newRoot string, cmdPath string, args []string) error {
	if newRoot == "" {
		return fmt.Errorf("se debe especificar la ruta del rootfs")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar")
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"child-init-chroot", newRoot, cmdPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = GetBasicSysProcAttr()

	return cmd.Run()
}

// RunPivotRoot lanza `cmdPath args...` aislando el sistema de archivos raíz mediante pivot_root.
func RunPivotRoot(newRoot string, cmdPath string, args []string) error {
	if newRoot == "" {
		return fmt.Errorf("se debe especificar la ruta del rootfs")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar")
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"child-init-pivot", newRoot, cmdPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = GetBasicSysProcAttr()

	return cmd.Run()
}

// ChildInit se ejecuta dentro del nuevo namespace (PID 1 del nuevo árbol).
// Configura el hostname, establece montajes privados, monta /proc y ejecuta el binario final.
func ChildInit(cmdPath string, args []string) error {
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar dentro del contenedor")
	}

	// 1. Configurar hostname en el UTS namespace
	if err := syscall.Sethostname([]byte("zerodaycontainer")); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	// 2. Asegurar que los montajes de este namespace sean privados (evita propagación al host)
	_ = syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")

	// 3. Montar /proc privado para reflejar los procesos del nuevo PID namespace
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc: %w", err)
	}

	// 4. Ejecutar el comando final solicitado por el usuario
	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ChildInitChroot se ejecuta dentro del nuevo namespace, aplica chroot sobre newRoot y ejecuta el comando.
func ChildInitChroot(newRoot string, cmdPath string, args []string) error {
	if newRoot == "" {
		return fmt.Errorf("se debe especificar la ruta del rootfs")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar dentro del contenedor")
	}

	// 1. Configurar hostname en el UTS namespace
	if err := syscall.Sethostname([]byte("zerodaycontainer")); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	// 2. Aplicar chroot a newRoot
	if err := rootfs.ApplyChroot(newRoot); err != nil {
		return fmt.Errorf("aplicar chroot: %w", err)
	}

	// 3. Asegurar montajes privados y montar /proc dentro del nuevo rootfs
	_ = syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc dentro del chroot: %w", err)
	}

	// 4. Ejecutar el comando final solicitado
	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ChildInitPivotRoot se ejecuta dentro del nuevo namespace, aplica pivot_root sobre newRoot y ejecuta el comando.
func ChildInitPivotRoot(newRoot string, cmdPath string, args []string) error {
	if newRoot == "" {
		return fmt.Errorf("se debe especificar la ruta del rootfs")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar dentro del contenedor")
	}

	// 1. Configurar hostname en el UTS namespace
	if err := syscall.Sethostname([]byte("zerodaycontainer")); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	// 2. Aplicar pivot_root a newRoot (desvincula la raíz vieja del host)
	if err := rootfs.ApplyPivotRoot(newRoot); err != nil {
		return fmt.Errorf("aplicar pivot_root: %w", err)
	}

	// 3. Montar /proc privado dentro del nuevo rootfs pivotado
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc dentro de pivot_root: %w", err)
	}

	// 4. Ejecutar el comando final solicitado
	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
