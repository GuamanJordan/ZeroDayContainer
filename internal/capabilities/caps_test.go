package capabilities

import (
	"runtime"
	"testing"
)

func TestGetDefaultDangerousCapabilities(t *testing.T) {
	caps := GetDefaultDangerousCapabilities()
	if runtime.GOOS == "linux" {
		if len(caps) == 0 {
			t.Error("se esperaba una lista no vacía de capabilities peligrosas en Linux")
		}
		foundSysAdmin := false
		for _, c := range caps {
			if c == CAP_SYS_ADMIN {
				foundSysAdmin = true
				break
			}
		}
		if !foundSysAdmin {
			t.Error("se esperaba encontrar CAP_SYS_ADMIN dentro de las capabilities peligrosas")
		}
	} else {
		if len(caps) != 0 {
			t.Errorf("se esperaba lista vacía en no-Linux, se obtuvo: %d", len(caps))
		}
	}
}
