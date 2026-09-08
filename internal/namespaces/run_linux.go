//go:build linux

package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/GuamanJordan/ZeroDayContainer/internal/capabilities"
	"github.com/GuamanJordan/ZeroDayContainer/internal/cgroups"
	"github.com/GuamanJordan/ZeroDayContainer/internal/network"
	"github.com/GuamanJordan/ZeroDayContainer/internal/rootfs"
)

// GetBasicSysProcAttr retorna la configuración SysProcAttr con las flags
// CLONE_NEWUTS, CLONE_NEWPID y CLONE_NEWNS activadas para aislamiento de hostname, PIDs y montajes.
func GetBasicSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
}

// GetNetworkSysProcAttr retorna la configuración SysProcAttr agregando CLONE_NEWNET para aislamiento de red.
func GetNetworkSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
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

// ContainerOpts define las opciones completas de aislamiento de un contenedor.
type ContainerOpts struct {
	ReadOnly  bool
	EnableNet bool
	Cgroups   cgroups.Config
}

// RunPivotRoot lanza `cmdPath args...` aislando el sistema de archivos raíz mediante pivot_root.
// Permite especificar si el filesystem del contenedor debe montarse en modo solo lectura (readOnly).
func RunPivotRoot(newRoot string, readOnly bool, cmdPath string, args []string) error {
	return RunPivotRootWithOptions(newRoot, ContainerOpts{ReadOnly: readOnly}, cmdPath, args)
}

// RunPivotRootWithCgroups lanza el contenedor aplicando pivot_root, modo solo lectura opcional y límites de cgroups v2.
func RunPivotRootWithCgroups(newRoot string, readOnly bool, cgCfg cgroups.Config, cmdPath string, args []string) error {
	return RunPivotRootWithOptions(newRoot, ContainerOpts{ReadOnly: readOnly, Cgroups: cgCfg}, cmdPath, args)
}

// RunPivotRootWithOptions lanza el contenedor aplicando las opciones de filesystem, red y cgroups especificadas.
func RunPivotRootWithOptions(newRoot string, opts ContainerOpts, cmdPath string, args []string) error {
	if newRoot == "" {
		return fmt.Errorf("se debe especificar la ruta del rootfs")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar")
	}

	initArgs := []string{"child-init-pivot"}
	if opts.ReadOnly {
		initArgs = append(initArgs, "--read-only")
	}
	initArgs = append(initArgs, newRoot, cmdPath)
	initArgs = append(initArgs, args...)

	cmd := exec.Command("/proc/self/exe", initArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if opts.EnableNet {
		cmd.SysProcAttr = GetNetworkSysProcAttr()
	} else {
		cmd.SysProcAttr = GetBasicSysProcAttr()
	}

	hasCgroups := opts.Cgroups.MemoryLimit != "" || opts.Cgroups.CPUs > 0 || opts.Cgroups.PIDsLimit > 0
	var cg *cgroups.Cgroup
	if hasCgroups {
		var err error
		cg, err = cgroups.New(fmt.Sprintf("zeroday-%d", time.Now().UnixNano()))
		if err != nil {
			return fmt.Errorf("error al crear cgroup: %w", err)
		}
		defer func() {
			_ = cg.Cleanup()
		}()

		if err := cg.ApplyLimits(opts.Cgroups); err != nil {
			return fmt.Errorf("error al aplicar límites de cgroup: %w", err)
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error al iniciar proceso contenedor: %w", err)
	}

	// Configurar cgroup si aplica
	if hasCgroups && cg != nil {
		if err := cg.AddProcess(cmd.Process.Pid); err != nil {
			_ = cmd.Process.Kill()
			return fmt.Errorf("error al asignar proceso a cgroup: %w", err)
		}
	}

	// Configurar red veth si se activó el flag
	if opts.EnableNet {
		netCfg := network.DefaultNetworkConfig(fmt.Sprintf("%d", cmd.Process.Pid))
		_ = network.SetupVethPair(cmd.Process.Pid, netCfg)
		defer func() {
			_ = network.CleanupVethPair(netCfg.HostVethName)
		}()
	}

	return cmd.Wait()
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

// ChildInitPivotRoot se ejecuta dentro del nuevo namespace, aplica pivot_root y monta /proc, /sys, /dev y /dev/pts.
// Aplica hardening como rootfs de solo lectura si readOnly es verdadero y asegura desmontajes limpios al salir.
func ChildInitPivotRoot(newRoot string, readOnly bool, cmdPath string, args []string) error {
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

	// 2. Aplicar pivot_root a newRoot con opciones de hardening
	if err := rootfs.ApplyPivotRootWithOptions(newRoot, rootfs.PivotOptions{ReadOnly: readOnly}); err != nil {
		return fmt.Errorf("aplicar pivot_root: %w", err)
	}

	// 3. Montar sistemas de archivos esenciales (/proc, /sys, /dev, /dev/pts) dentro del rootfs pivotado
	if err := rootfs.MountEssentialFilesystems(); err != nil {
		return fmt.Errorf("montar sistemas de archivos esenciales: %w", err)
	}
	defer func() {
		_ = rootfs.UnmountEssentialFilesystems()
	}()

	// 4. Reducir capabilities peligrosas y activar NO_NEW_PRIVS para el proceso contenedor
	_ = capabilities.RestrictPrivileges()

	// 5. Ejecutar el comando final solicitado
	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
