//go:build integration

package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestIntegrationNamespaces valida que mc run-basic ejecute el aislamiento real de UTS y PID namespaces en Linux.
func TestIntegrationNamespaces(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Saltando prueba de integración: se requieren privilegios de root (euid == 0)")
	}

	// Construir el binario mc antes de probarlo
	buildCmd := exec.Command("go", "build", "-o", "/tmp/mc_test", "../../cmd/mc")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Falló la compilación del binario mc: %v\nSalida: %s", err, string(out))
	}
	defer os.Remove("/tmp/mc_test")

	// Probar hostname aislado dentro del UTS namespace
	cmd := exec.Command("/tmp/mc_test", "run-basic", "hostname")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Falló la ejecución de mc run-basic hostname: %v\nSalida: %s", err, string(out))
	}

	output := strings.TrimSpace(string(out))
	expectedHost := "zerodaycontainer"
	if output != expectedHost {
		t.Errorf("Se esperaba hostname '%s', se obtuvo: '%s'", expectedHost, output)
	}
}
