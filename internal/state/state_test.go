package state

import (
	"os"
	"testing"
	"time"
)

func TestContainerStateLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mc-state-test-*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	SetStateDir(tempDir)

	info := ContainerInfo{
		ID:        "test-cont-1",
		PID:       os.Getpid(),
		Command:   "/bin/sh",
		Rootfs:    "/tmp/alpine",
		CreatedAt: time.Now(),
		Status:    "running",
	}

	if err := SaveContainer(info); err != nil {
		t.Fatalf("error guardando estado: %v", err)
	}

	list, err := ListContainers()
	if err != nil {
		t.Fatalf("error listando contenedores: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("se esperaba 1 contenedor, se obtuvieron: %d", len(list))
	}
	if list[0].ID != "test-cont-1" {
		t.Errorf("ID esperado 'test-cont-1', obtenido '%s'", list[0].ID)
	}

	if err := RemoveContainer("test-cont-1"); err != nil {
		t.Fatalf("error eliminando estado: %v", err)
	}

	listAfter, err := ListContainers()
	if err != nil {
		t.Fatalf("error listando tras eliminar: %v", err)
	}
	if len(listAfter) != 0 {
		t.Fatalf("se esperaban 0 contenedores tras eliminar, se obtuvieron: %d", len(listAfter))
	}
}
