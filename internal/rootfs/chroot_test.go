package rootfs

import (
	"testing"
)

func TestApplyChrootValidation(t *testing.T) {
	err := ApplyChroot("")
	if err == nil {
		t.Error("se esperaba un error al pasar newRoot vacío")
	}

	err = ApplyChroot("/ruta/inexistente/zerodaycontainer")
	if err == nil {
		t.Error("se esperaba un error al pasar una ruta inexistente")
	}
}
