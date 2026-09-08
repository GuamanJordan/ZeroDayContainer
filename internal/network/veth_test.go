package network

import (
	"strings"
	"testing"
)

func TestDefaultNetworkConfig(t *testing.T) {
	cfg := DefaultNetworkConfig("cont12345")
	if !strings.HasPrefix(cfg.HostVethName, "veth-") {
		t.Errorf("HostVethName inesperado: %s", cfg.HostVethName)
	}
	if !strings.HasPrefix(cfg.GuestVethName, "vethg-") {
		t.Errorf("GuestVethName inesperado: %s", cfg.GuestVethName)
	}
	if cfg.HostIP != "10.16.8.1/24" {
		t.Errorf("HostIP inesperado: %s", cfg.HostIP)
	}
	if cfg.GuestIP != "10.16.8.2/24" {
		t.Errorf("GuestIP inesperado: %s", cfg.GuestIP)
	}
}

func TestCleanupVethPairEmpty(t *testing.T) {
	if err := CleanupVethPair(""); err != nil {
		t.Errorf("CleanupVethPair('') no debería retornar error: %v", err)
	}
}
