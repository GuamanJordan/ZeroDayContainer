package namespaces

import (
	"testing"
)

func TestExecInContainerValidation(t *testing.T) {
	// PID inválido
	err := ExecInContainer(0, "/bin/sh", nil)
	if err == nil {
		t.Error("se esperaba error con PID <= 0")
	}

	err = ExecInContainer(-5, "/bin/sh", nil)
	if err == nil {
		t.Error("se esperaba error con PID negativo")
	}

	// Comando vacío
	err = ExecInContainer(1234, "", nil)
	if err == nil {
		t.Error("se esperaba error con comando vacío")
	}
}
