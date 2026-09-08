//go:build !linux

package network

import "fmt"

// PortMapping define la redirección de un puerto del host hacia un puerto del contenedor.
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

// ParsePortMapping en sistemas no-Linux retorna error.
func ParsePortMapping(spec string) (PortMapping, error) {
	return PortMapping{}, fmt.Errorf("port forwarding solo está soportado en Linux")
}

// SetupPortForwarding en sistemas no-Linux retorna error.
func SetupPortForwarding(mapping PortMapping, containerIP string) error {
	return fmt.Errorf("port forwarding solo está soportado en Linux")
}

// CleanupPortForwarding en sistemas no-Linux es no-op.
func CleanupPortForwarding(mapping PortMapping, containerIP string) error {
	return nil
}
