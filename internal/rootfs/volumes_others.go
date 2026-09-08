//go:build !linux

package rootfs

import "fmt"

// VolumeMount define la configuración de un bind mount entre host y contenedor.
type VolumeMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// ParseVolumeSpec analiza una especificación de volumen en formato /host:/contenedor[:ro|rw].
func ParseVolumeSpec(spec string) (VolumeMount, error) {
	return VolumeMount{}, fmt.Errorf("bind mounts solo están soportados en Linux")
}

// MountVolumes en sistemas no-Linux retorna error.
func MountVolumes(rootfs string, volumes []VolumeMount) ([]string, error) {
	return nil, fmt.Errorf("bind mounts solo están soportados en Linux")
}

// UnmountVolumes en sistemas no-Linux es no-op.
func UnmountVolumes(targets []string) {}
