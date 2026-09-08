package network

import (
	"testing"
)

func TestParsePortMapping(t *testing.T) {
	// 1. Caso estándar tcp
	pm, err := ParsePortMapping("8080:80")
	if err != nil {
		t.Fatalf("ParsePortMapping falló: %v", err)
	}
	if pm.HostPort != 8080 || pm.ContainerPort != 80 || pm.Protocol != "tcp" {
		t.Errorf("PortMapping inesperado: %+v", pm)
	}

	// 2. Caso explícito udp
	pm2, err := ParsePortMapping("5353:53/udp")
	if err != nil {
		t.Fatalf("ParsePortMapping con udp falló: %v", err)
	}
	if pm2.HostPort != 5353 || pm2.ContainerPort != 53 || pm2.Protocol != "udp" {
		t.Errorf("PortMapping udp inesperado: %+v", pm2)
	}

	// 3. Caso explícito tcp
	pm3, err := ParsePortMapping("4433:443/tcp")
	if err != nil {
		t.Fatalf("ParsePortMapping con tcp falló: %v", err)
	}
	if pm3.HostPort != 4433 || pm3.ContainerPort != 443 || pm3.Protocol != "tcp" {
		t.Errorf("PortMapping tcp inesperado: %+v", pm3)
	}

	// 4. Casos inválidos
	invalid := []string{
		"",
		"8080",
		":80",
		"8080:",
		"abc:80",
		"8080:abc",
		"0:80",
		"8080:0",
		"70000:80",
		"8080:70000",
		"8080:80/sctp",
		"8080:80/tcp/extra",
	}

	for _, spec := range invalid {
		_, err := ParsePortMapping(spec)
		if err == nil {
			t.Errorf("se esperaba error para spec inválido %q", spec)
		}
	}
}
