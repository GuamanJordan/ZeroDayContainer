package rootfs

import (
	"testing"
)

func TestApplyPivotRootValidation(t *testing.T) {
	err := ApplyPivotRoot("")
	if err == nil {
		t.Error("se esperaba un error al pasar newRoot vacío")
	}

	err = ApplyPivotRoot("/ruta/inexistente/zerodaycontainer")
	if err == nil {
		t.Error("se esperaba un error al pasar una ruta inexistente")
	}
}
