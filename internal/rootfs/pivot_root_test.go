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

	err = ApplyPivotRootWithOptions("", PivotOptions{ReadOnly: true})
	if err == nil {
		t.Error("se esperaba un error al pasar newRoot vacío en ApplyPivotRootWithOptions")
	}

	err = ApplyPivotRootWithOptions("/ruta/inexistente/zerodaycontainer", PivotOptions{ReadOnly: true})
	if err == nil {
		t.Error("se esperaba un error al pasar una ruta inexistente en ApplyPivotRootWithOptions")
	}
}

func TestRemountReadOnlyValidation(t *testing.T) {
	err := RemountReadOnly("")
	if err == nil {
		t.Error("se esperaba un error al pasar un target vacío a RemountReadOnly")
	}
}
