//go:build linux

package network

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PortMapping define la redirección de un puerto del host hacia un puerto del contenedor.
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

// ParsePortMapping analiza una especificación de puerto en formato hostPort:containerPort[/protocol].
func ParsePortMapping(spec string) (PortMapping, error) {
	if spec == "" {
		return PortMapping{}, fmt.Errorf("especificación de puerto vacía")
	}

	protocol := "tcp"
	protoParts := strings.Split(spec, "/")
	if len(protoParts) == 2 {
		proto := strings.ToLower(protoParts[1])
		if proto != "tcp" && proto != "udp" {
			return PortMapping{}, fmt.Errorf("protocolo inválido %q (se espera 'tcp' o 'udp')", proto)
		}
		protocol = proto
		spec = protoParts[0]
	} else if len(protoParts) > 2 {
		return PortMapping{}, fmt.Errorf("formato de puerto inválido %q", spec)
	}

	ports := strings.Split(spec, ":")
	if len(ports) != 2 {
		return PortMapping{}, fmt.Errorf("se espera formato hostPort:containerPort en %q", spec)
	}

	hPort, err := strconv.Atoi(ports[0])
	if err != nil || hPort < 1 || hPort > 65535 {
		return PortMapping{}, fmt.Errorf("puerto del host inválido %q (rango 1-65535)", ports[0])
	}

	cPort, err := strconv.Atoi(ports[1])
	if err != nil || cPort < 1 || cPort > 65535 {
		return PortMapping{}, fmt.Errorf("puerto del contenedor inválido %q (rango 1-65535)", ports[1])
	}

	return PortMapping{
		HostPort:      hPort,
		ContainerPort: cPort,
		Protocol:      protocol,
	}, nil
}

// SetupPortForwarding configura las reglas iptables DNAT en PREROUTING y OUTPUT.
func SetupPortForwarding(mapping PortMapping, containerIP string) error {
	if containerIP == "" {
		return fmt.Errorf("IP del contenedor requerida para reenvío de puertos")
	}

	ipOnly := strings.Split(containerIP, "/")[0]
	dest := fmt.Sprintf("%s:%d", ipOnly, mapping.ContainerPort)
	hPortStr := strconv.Itoa(mapping.HostPort)

	// 1. Regla en PREROUTING (tráfico externo hacia el host)
	cmdPre := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", mapping.Protocol,
		"--dport", hPortStr, "-j", "DNAT", "--to-destination", dest)
	if out, err := cmdPre.CombinedOutput(); err != nil {
		return fmt.Errorf("iptables DNAT PREROUTING: %s (%w)", string(out), err)
	}

	// 2. Regla en OUTPUT (tráfico local desde el propio host hacia 127.0.0.1)
	cmdOut := exec.Command("iptables", "-t", "nat", "-A", "OUTPUT", "-p", mapping.Protocol,
		"-d", "127.0.0.1", "--dport", hPortStr, "-j", "DNAT", "--to-destination", dest)
	_ = cmdOut.Run()

	return nil
}

// CleanupPortForwarding elimina las reglas iptables configuradas para la redirección.
func CleanupPortForwarding(mapping PortMapping, containerIP string) error {
	ipOnly := strings.Split(containerIP, "/")[0]
	dest := fmt.Sprintf("%s:%d", ipOnly, mapping.ContainerPort)
	hPortStr := strconv.Itoa(mapping.HostPort)

	cmdPre := exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", mapping.Protocol,
		"--dport", hPortStr, "-j", "DNAT", "--to-destination", dest)
	_ = cmdPre.Run()

	cmdOut := exec.Command("iptables", "-t", "nat", "-D", "OUTPUT", "-p", mapping.Protocol,
		"-d", "127.0.0.1", "--dport", hPortStr, "-j", "DNAT", "--to-destination", dest)
	_ = cmdOut.Run()

	return nil
}
