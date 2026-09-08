package rootfs

import (
	"path/filepath"
	"testing"
)

func TestParseVolumeSpec(t *testing.T) {
	// 1. Caso estándar lectura y escritura
	v1, err := ParseVolumeSpec("/tmp/data:/app/data")
	if err != nil {
		t.Fatalf("ParseVolumeSpec falló: %v", err)
	}
	if v1.ContainerPath != "/app/data" {
		t.Errorf("ContainerPath esperado '/app/data', obtenido %q", v1.ContainerPath)
	}
	if v1.ReadOnly {
		t.Errorf("ReadOnly esperado false, obtenido true")
	}

	// 2. Caso modo solo lectura :ro
	v2, err := ParseVolumeSpec("/tmp/config:/etc/app:ro")
	if err != nil {
		t.Fatalf("ParseVolumeSpec con :ro falló: %v", err)
	}
	if !v2.ReadOnly {
		t.Errorf("ReadOnly esperado true, obtenido false")
	}

	// 3. Caso explícito :rw
	v3, err := ParseVolumeSpec("/tmp/logs:/var/log:rw")
	if err != nil {
		t.Fatalf("ParseVolumeSpec con :rw falló: %v", err)
	}
	if v3.ReadOnly {
		t.Errorf("ReadOnly esperado false, obtenido true")
	}

	// 4. Ruta relativa del host se convierte en absoluta
	v4, err := ParseVolumeSpec("./data:/workspace")
	if err != nil {
		t.Fatalf("ParseVolumeSpec con ruta relativa falló: %v", err)
	}
	if !filepath.IsAbs(v4.HostPath) {
		t.Errorf("HostPath no es absoluto: %q", v4.HostPath)
	}

	// 5. Casos inválidos
	invalidSpecs := []string{
		"",
		"/singlepath",
		":/container",
		"/host:",
		"/host:relative/path",
		"/host:/container:invalidmode",
		"/host:/container:ro:extra",
	}

	for _, spec := range invalidSpecs {
		_, err := ParseVolumeSpec(spec)
		if err == nil {
			t.Errorf("se esperaba error para spec inválido %q", spec)
		}
	}
}
