//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// MountSpec define la estructura para especificar un montaje pseudo o de sistema de archivos.
type MountSpec struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}

// GetEssentialMounts retorna la lista ordenada de montajes esenciales para un contenedor (/proc, /sys, /dev, /dev/pts).
func GetEssentialMounts() []MountSpec {
	return []MountSpec{
		{Source: "proc", Target: "/proc", FSType: "proc", Flags: 0, Data: ""},
		{Source: "sysfs", Target: "/sys", FSType: "sysfs", Flags: 0, Data: ""},
		{Source: "tmpfs", Target: "/dev", FSType: "tmpfs", Flags: syscall.MS_NOSUID | syscall.MS_STRICTATIME, Data: "mode=755"},
		{Source: "devpts", Target: "/dev/pts", FSType: "devpts", Flags: 0, Data: "newinstance,ptmxmode=0666,mode=0620,gid=5"},
	}
}

// MountEssentialFilesystems crea los directorios destino y monta los sistemas de archivos virtuales
// indispensables dentro del contenedor (/proc, /sys, /dev y /dev/pts).
// Si ocurre un error durante el proceso de montaje, realiza un rollback desmontando en orden
// inverso todos los puntos montados previamente para evitar filtraciones de recursos.
func MountEssentialFilesystems() (err error) {
	var mounted []string
	defer func() {
		if err != nil {
			RollbackMounts(mounted)
		}
	}()

	for _, m := range GetEssentialMounts() {
		if err := os.MkdirAll(m.Target, 0750); err != nil {
			return fmt.Errorf("mkdir '%s': %w", m.Target, err)
		}
		if err := syscall.Mount(m.Source, m.Target, m.FSType, m.Flags, m.Data); err != nil {
			return fmt.Errorf("mount %s en %s (%s): %w", m.Source, m.Target, m.FSType, err)
		}
		mounted = append(mounted, m.Target)
	}

	if err := createEssentialDevNodes(); err != nil {
		return fmt.Errorf("crear nodos en /dev: %w", err)
	}

	return nil
}

// RollbackMounts desmonta en orden inverso los puntos de montaje indicados usando MNT_DETACH.
func RollbackMounts(targets []string) {
	for i := len(targets) - 1; i >= 0; i-- {
		target := targets[i]
		_ = syscall.Unmount(target, syscall.MNT_DETACH)
	}
}

// UnmountEssentialFilesystems desmonta de forma segura y en orden inverso los sistemas de archivos esenciales.
func UnmountEssentialFilesystems() error {
	mounts := GetEssentialMounts()
	var errs []error
	for i := len(mounts) - 1; i >= 0; i-- {
		target := mounts[i].Target
		if err := syscall.Unmount(target, syscall.MNT_DETACH); err != nil {
			errs = append(errs, fmt.Errorf("desmontar '%s': %w", target, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errores al desmontar sistemas de archivos esenciales: %v", errs)
	}
	return nil
}

// createEssentialDevNodes crea los enlaces simbólicos y nodos básicos en /dev.
func createEssentialDevNodes() error {
	devices := []string{"null", "zero", "random", "urandom", "tty"}

	for _, dev := range devices {
		targetDev := filepath.Join("/dev", dev)
		f, err := os.Create(targetDev)
		if err == nil {
			_ = f.Close()
			_ = syscall.Mount(filepath.Join("/proc/kcore"), targetDev, "", syscall.MS_BIND, "")
		}
	}

	// Enlaces simbólicos para descriptores estándar
	_ = os.Symlink("/proc/self/fd", "/dev/fd")
	_ = os.Symlink("/proc/self/fd/0", "/dev/stdin")
	_ = os.Symlink("/proc/self/fd/1", "/dev/stdout")
	_ = os.Symlink("/proc/self/fd/2", "/dev/stderr")
	_ = os.Symlink("/dev/pts/ptmx", "/dev/ptmx")

	return nil
}
