//go:build linux

package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// VolumeMount define la configuración de un bind mount entre host y contenedor.
type VolumeMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// ParseVolumeSpec analiza una especificación de volumen en formato /host:/contenedor[:ro|rw].
func ParseVolumeSpec(spec string) (VolumeMount, error) {
	if spec == "" {
		return VolumeMount{}, fmt.Errorf("especificación de volumen vacía")
	}

	parts := strings.Split(spec, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return VolumeMount{}, fmt.Errorf("formato inválido de volumen %q (se espera /host:/contenedor[:ro|rw])", spec)
	}

	hostPath := parts[0]
	containerPath := parts[1]

	if hostPath == "" || containerPath == "" {
		return VolumeMount{}, fmt.Errorf("las rutas de host y contenedor no pueden estar vacías en %q", spec)
	}

	if !filepath.IsAbs(containerPath) {
		return VolumeMount{}, fmt.Errorf("la ruta del contenedor debe ser absoluta: %q", containerPath)
	}

	absHostPath, err := filepath.Abs(hostPath)
	if err != nil {
		return VolumeMount{}, fmt.Errorf("resolver ruta del host %q: %w", hostPath, err)
	}

	readOnly := false
	if len(parts) == 3 {
		mode := strings.ToLower(parts[2])
		switch mode {
		case "ro":
			readOnly = true
		case "rw":
			readOnly = false
		default:
			return VolumeMount{}, fmt.Errorf("modo de montaje inválido %q (se espera 'ro' o 'rw')", mode)
		}
	}

	return VolumeMount{
		HostPath:      absHostPath,
		ContainerPath: containerPath,
		ReadOnly:      readOnly,
	}, nil
}

// MountVolumes realiza los bind mounts de los volúmenes especificados sobre el rootfs destino.
// Retorna la lista de rutas montadas para su posterior desmontaje seguro.
func MountVolumes(rootfs string, volumes []VolumeMount) ([]string, error) {
	var mounted []string

	for _, vol := range volumes {
		// Verificar que la ruta en el host exista
		hostInfo, err := os.Stat(vol.HostPath)
		if err != nil {
			UnmountVolumes(mounted)
			return nil, fmt.Errorf("la ruta del host %q no existe: %w", vol.HostPath, err)
		}

		targetPath := filepath.Join(rootfs, vol.ContainerPath)

		if hostInfo.IsDir() {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				UnmountVolumes(mounted)
				return nil, fmt.Errorf("crear punto de montaje en directorio %q: %w", targetPath, err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
				UnmountVolumes(mounted)
				return nil, fmt.Errorf("crear directorio padre para archivo %q: %w", targetPath, err)
			}
			f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDONLY, 0600)
			if err != nil {
				UnmountVolumes(mounted)
				return nil, fmt.Errorf("crear archivo destino para bind mount %q: %w", targetPath, err)
			}
			_ = f.Close()
		}

		// 1. Realizar bind mount
		flags := uintptr(syscall.MS_BIND | syscall.MS_REC)
		if err := syscall.Mount(vol.HostPath, targetPath, "", flags, ""); err != nil {
			UnmountVolumes(mounted)
			return nil, fmt.Errorf("bind mount de %q en %q: %w", vol.HostPath, targetPath, err)
		}
		mounted = append(mounted, targetPath)

		// 2. Si se solicitó solo lectura, aplicar remount con MS_RDONLY
		if vol.ReadOnly {
			remountFlags := uintptr(syscall.MS_BIND | syscall.MS_REC | syscall.MS_REMOUNT | syscall.MS_RDONLY)
			if err := syscall.Mount("", targetPath, "", remountFlags, ""); err != nil {
				UnmountVolumes(mounted)
				return nil, fmt.Errorf("remount read-only en %q: %w", targetPath, err)
			}
		}
	}

	return mounted, nil
}

// UnmountVolumes desmonta de forma segura y en orden inverso los volúmenes montados.
func UnmountVolumes(targets []string) {
	for i := len(targets) - 1; i >= 0; i-- {
		target := targets[i]
		_ = syscall.Unmount(target, syscall.MNT_DETACH)
	}
}
