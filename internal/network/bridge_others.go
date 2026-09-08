//go:build !linux

package network

import "fmt"

const (
	DefaultBridgeName = "mc0"
	DefaultBridgeIP   = "10.16.8.1/24"
	DefaultSubnet     = "10.16.8.0/24"
	DefaultGatewayIP  = "10.16.8.1"
)

// SetupBridge en sistemas no-Linux retorna error de incompatibilidad.
func SetupBridge(bridgeName, bridgeIP string) error {
	return fmt.Errorf("bridges virtuales solo están soportados en Linux")
}

// AttachToBridge en sistemas no-Linux retorna error de incompatibilidad.
func AttachToBridge(hostVeth, bridgeName string) error {
	return fmt.Errorf("bridges virtuales solo están soportados en Linux")
}

// EnableNAT en sistemas no-Linux retorna error de incompatibilidad.
func EnableNAT(subnet, bridgeName string) error {
	return fmt.Errorf("NAT e iptables solo están soportados en Linux")
}

// SetupDefaultGateway en sistemas no-Linux retorna error de incompatibilidad.
func SetupDefaultGateway(pid int, gatewayIP string) error {
	return fmt.Errorf("enrutamiento en netns solo está soportado en Linux")
}
