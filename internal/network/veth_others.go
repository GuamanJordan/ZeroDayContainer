//go:build !linux

package network

import "fmt"

// NetworkConfig contiene la configuración de red para el par veth del contenedor.
type NetworkConfig struct {
	HostVethName  string
	GuestVethName string
	HostIP        string
	GuestIP       string
}

// DefaultNetworkConfig retorna la configuración de red por defecto.
func DefaultNetworkConfig(containerID string) NetworkConfig {
	return NetworkConfig{}
}

// SetupVethPair en sistemas no-Linux retorna error de incompatibilidad.
func SetupVethPair(pid int, cfg NetworkConfig) error {
	return fmt.Errorf("network namespaces y veth pairs solo están soportados en Linux")
}

// CleanupVethPair en sistemas no-Linux es un no-op.
func CleanupVethPair(hostVeth string) error {
	return nil
}
