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

// TestIntegrationReadOnlyRootfs valida la ejecución de un contenedor con opción --read-only.
func TestIntegrationReadOnlyRootfs(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Saltando prueba de integración: se requieren privilegios de root (euid == 0)")
	}

	buildCmd := exec.Command("go", "build", "-o", "/tmp/mc_test_ro", "../../cmd/mc")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Falló la compilación del binario mc: %v\nSalida: %s", err, string(out))
	}
	defer os.Remove("/tmp/mc_test_ro")

	// Crear un directorio temporal para simular un rootfs mínimo
	tempRootfs, err := os.MkdirTemp("", "rootfs-ro-test-*")
	if err != nil {
		t.Fatalf("Error creando rootfs temporal: %v", err)
	}
	defer os.RemoveAll(tempRootfs)

	// Ejecutar mc run --read-only sobre el directorio temporal con un comando inexistente
	// para validar que la fase de preparación, pivot_root y desmontaje seguro finalizan sin panics
	cmd := exec.Command("/tmp/mc_test_ro", "run", "--read-only", tempRootfs, "/bin/inexistente")
	out, _ := cmd.CombinedOutput()
	output := string(out)

	if strings.Contains(output, "panic") {
		t.Errorf("La ejecución con --read-only causó un panic inesperado: %s", output)
	}
}
