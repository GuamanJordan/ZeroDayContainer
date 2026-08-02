package namespaces

import (
	"runtime"
	"syscall"
	"testing"
)

func TestGetBasicSysProcAttr(t *testing.T) {
	attr := GetBasicSysProcAttr()
	if attr == nil {
		t.Fatal("GetBasicSysProcAttr() retornó nil")
	}

	if runtime.GOOS == "linux" {
		expected := uintptr(syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS)
		if attr.Cloneflags != expected {
			t.Errorf("se esperaba Cloneflags=%v, se obtuvo=%v", expected, attr.Cloneflags)
		}
	}
}

func TestRunBasicValidation(t *testing.T) {
	err := RunBasic("", nil)
	if err == nil {
		t.Error("se esperaba un error al pasar cmdPath vacío")
	}
}

func TestChildInitValidation(t *testing.T) {
	err := ChildInit("", nil)
	if err == nil {
		t.Error("se esperaba un error al pasar cmdPath vacío")
	}
}
