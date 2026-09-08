package rootfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayConfigValidation(t *testing.T) {
	// lower vacío
	err := MountOverlay(OverlayConfig{UpperDir: "/tmp/u", WorkDir: "/tmp/w", MergedDir: "/tmp/m"})
	if err == nil {
		t.Error("se esperaba error con LowerDir vacío")
	}

	// upper vacío
	err = MountOverlay(OverlayConfig{LowerDir: "/tmp/l", WorkDir: "/tmp/w", MergedDir: "/tmp/m"})
	if err == nil {
		t.Error("se esperaba error con UpperDir vacío")
	}

	// work vacío
	err = MountOverlay(OverlayConfig{LowerDir: "/tmp/l", UpperDir: "/tmp/u", MergedDir: "/tmp/m"})
	if err == nil {
		t.Error("se esperaba error con WorkDir vacío")
	}

	// unmount vacío
	err = UnmountOverlay("")
	if err == nil {
		t.Error("se esperaba error al desmontar mergedDir vacío")
	}
}

func TestSetupOverlayDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := OverlayConfig{
		LowerDir:  tmpDir,
		UpperDir:  filepath.Join(tmpDir, "upper"),
		WorkDir:   filepath.Join(tmpDir, "work"),
		MergedDir: filepath.Join(tmpDir, "merged"),
	}

	if err := SetupOverlayDirectories(cfg); err != nil {
		t.Fatalf("SetupOverlayDirectories falló: %v", err)
	}

	for _, d := range []string{cfg.UpperDir, cfg.WorkDir, cfg.MergedDir} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("directorio %s no fue creado", d)
		}
	}
}
