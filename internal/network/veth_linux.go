//go:build linux

package network

import (
	"fmt"
	"os/exec"
	"strconv"
)

// NetworkConfig contiene la configuración de red para el par veth del contenedor.
type NetworkConfig struct {
	HostVethName  string
	GuestVethName string
	HostIP        string // ej: 10.16.8.1/24
	GuestIP       string // ej: 10.16.8.2/24
}

// DefaultNetworkConfig retorna la configuración de red por defecto para aislamiento básico.
func DefaultNetworkConfig(containerID string) NetworkConfig {
	if len(containerID) > 5 {
		containerID = containerID[:5]
	}
	return NetworkConfig{
		HostVethName:  fmt.Sprintf("veth-%s", containerID),
		GuestVethName: fmt.Sprintf("vethg-%s", containerID),
		HostIP:        "10.16.8.1/24",
		GuestIP:       "10.16.8.2/24",
	}
}

// SetupVethPair crea un par de interfaces veth, mueve el extremo invitado al namespace del PID
// y asigna las IPs configuradas a ambos extremos.
func SetupVethPair(pid int, cfg NetworkConfig) error {
	// 1. Crear el par veth en el host
	cmd := exec.Command("ip", "link", "add", cfg.HostVethName, "type", "veth", "peer", "name", cfg.GuestVethName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ip link add veth: %s (%w)", string(out), err)
	}

	// 2. Mover el extremo guest al netns del PID
	cmd = exec.Command("ip", "link", "set", cfg.GuestVethName, "netns", strconv.Itoa(pid))
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = CleanupVethPair(cfg.HostVethName)
		return fmt.Errorf("ip link set guest netns: %s (%w)", string(out), err)
	}

	// 3. Configurar IP en el host y levantar interfaz
	cmd = exec.Command("ip", "addr", "add", cfg.HostIP, "dev", cfg.HostVethName)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = CleanupVethPair(cfg.HostVethName)
		return fmt.Errorf("ip addr add host: %s (%w)", string(out), err)
	}
	cmd = exec.Command("ip", "link", "set", cfg.HostVethName, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = CleanupVethPair(cfg.HostVethName)
		return fmt.Errorf("ip link set host up: %s (%w)", string(out), err)
	}

	// 4. Configurar IP dentro del contenedor mediante nsenter
	cmd = exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n", "ip", "addr", "add", cfg.GuestIP, "dev", cfg.GuestVethName)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = CleanupVethPair(cfg.HostVethName)
		return fmt.Errorf("nsenter add guest ip: %s (%w)", string(out), err)
	}
	cmd = exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n", "ip", "link", "set", cfg.GuestVethName, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = CleanupVethPair(cfg.HostVethName)
		return fmt.Errorf("nsenter set guest up: %s (%w)", string(out), err)
	}
	cmd = exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n", "ip", "link", "set", "lo", "up")
	_ = cmd.Run()

	return nil
}

// CleanupVethPair elimina la interfaz veth del host (lo cual destruye automáticamente el par).
func CleanupVethPair(hostVeth string) error {
	if hostVeth == "" {
		return nil
	}
	cmd := exec.Command("ip", "link", "delete", hostVeth)
	return cmd.Run()
}
