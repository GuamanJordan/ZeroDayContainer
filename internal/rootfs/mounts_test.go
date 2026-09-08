package rootfs

import (
	"runtime"
	"testing"
)

func TestGetEssentialMounts(t *testing.T) {
	mounts := GetEssentialMounts()

	if runtime.GOOS == "linux" {
		if len(mounts) != 4 {
			t.Fatalf("se esperaban 4 montajes esenciales, se obtuvieron: %d", len(mounts))
		}
		expectedTargets := map[string]string{
			"/proc":    "proc",
			"/sys":     "sysfs",
			"/dev":     "tmpfs",
			"/dev/pts": "devpts",
		}
		for _, m := range mounts {
			fsType, exists := expectedTargets[m.Target]
			if !exists {
				t.Errorf("Target de montaje inesperado: %s", m.Target)
			}
			if m.FSType != fsType {
				t.Errorf("para target %s se esperaba FSType=%s, se obtuvo=%s", m.Target, fsType, m.FSType)
			}
		}
	} else {
		if len(mounts) != 0 {
			t.Errorf("en plataformas no-Linux se esperaba lista vacía, se obtuvieron: %d", len(mounts))
		}
	}
}

func TestRollbackMountsNoPanic(t *testing.T) {
	// Verificar que RollbackMounts maneje de forma segura slices nulos o vacíos sin causar panic
	RollbackMounts(nil)
	RollbackMounts([]string{})
	RollbackMounts([]string{"/nonexistent/test/path"})
}
