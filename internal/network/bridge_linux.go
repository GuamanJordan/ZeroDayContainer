//go:build linux

package network

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

const (
	// DefaultBridgeName es el nombre por defecto del puente virtual.
	DefaultBridgeName = "mc0"
	// DefaultBridgeIP es la IP y máscara asignadas a la interfaz bridge en el host.
	DefaultBridgeIP = "10.16.8.1/24"
	// DefaultSubnet es la subred administrada por el puente virtual.
	DefaultSubnet = "10.16.8.0/24"
	// DefaultGatewayIP es la dirección IP de salida para los contenedores.
	DefaultGatewayIP = "10.16.8.1"
)

// SetupBridge crea la interfaz bridge virtual mc0 y le asigna la IP del gateway si no existe.
func SetupBridge(bridgeName, bridgeIP string) error {
	if bridgeName == "" {
		bridgeName = DefaultBridgeName
	}
	if bridgeIP == "" {
		bridgeIP = DefaultBridgeIP
	}

	// Verificar si el bridge ya existe
	if err := exec.Command("ip", "link", "show", bridgeName).Run(); err == nil {
		_ = exec.Command("ip", "link", "set", bridgeName, "up").Run()
		return nil
	}

	// 1. Crear el bridge
	cmd := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("crear bridge '%s': %s (%w)", bridgeName, string(out), err)
	}

	// 2. Asignar IP al bridge
	cmd = exec.Command("ip", "addr", "add", bridgeIP, "dev", bridgeName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("asignar IP '%s' a bridge: %s (%w)", bridgeIP, string(out), err)
	}

	// 3. Levantar interfaz bridge
	cmd = exec.Command("ip", "link", "set", bridgeName, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("levantar bridge '%s': %s (%w)", bridgeName, string(out), err)
	}

	return nil
}

// AttachToBridge conecta el extremo host de una veth al bridge indicado.
func AttachToBridge(hostVeth, bridgeName string) error {
	if hostVeth == "" || bridgeName == "" {
		return fmt.Errorf("nombres de veth y bridge requeridos")
	}
	cmd := exec.Command("ip", "link", "set", hostVeth, "master", bridgeName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("conectar '%s' al bridge '%s': %s (%w)", hostVeth, bridgeName, string(out), err)
	}
	return nil
}

// EnableNAT activa el reenvío de paquetes IPv4 en el kernel y agrega la regla iptables MASQUERADE.
func EnableNAT(subnet, bridgeName string) error {
	if subnet == "" {
		subnet = DefaultSubnet
	}
	if bridgeName == "" {
		bridgeName = DefaultBridgeName
	}

	// 1. Habilitar ip_forward en el kernel
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0600); err != nil {
		_ = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()
	}

	// 2. Comprobar si la regla MASQUERADE ya existe
	checkCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE")
	if checkCmd.Run() == nil {
		return nil
	}

	// 3. Añadir regla MASQUERADE
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("iptables MASQUERADE: %s (%w)", string(out), err)
	}

	return nil
}

// SetupDefaultGateway configura la ruta por defecto (default gateway) en el netns del contenedor.
func SetupDefaultGateway(pid int, gatewayIP string) error {
	if gatewayIP == "" {
		gatewayIP = DefaultGatewayIP
	}
	cmd := exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n", "ip", "route", "add", "default", "via", gatewayIP)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nsenter add default route: %s (%w)", string(out), err)
	}
	return nil
}
