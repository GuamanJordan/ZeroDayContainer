package network

import (
	"testing"
)

func TestBridgeConstants(t *testing.T) {
	if DefaultBridgeName != "mc0" {
		t.Errorf("nombre de bridge esperado 'mc0', obtenido '%s'", DefaultBridgeName)
	}
	if DefaultBridgeIP != "10.16.8.1/24" {
		t.Errorf("IP de bridge esperada '10.16.8.1/24', obtenida '%s'", DefaultBridgeIP)
	}
	if DefaultGatewayIP != "10.16.8.1" {
		t.Errorf("Gateway esperado '10.16.8.1', obtenido '%s'", DefaultGatewayIP)
	}
}
