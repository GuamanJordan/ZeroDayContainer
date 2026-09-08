package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var defaultStateDir = "/tmp/mc/containers"

// ContainerInfo representa el estado persistido de un contenedor en ejecución.
type ContainerInfo struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	Rootfs    string    `json:"rootfs"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

// SetStateDir permite configurar el directorio de estado (útil para pruebas).
func SetStateDir(dir string) {
	defaultStateDir = dir
}

// SaveContainer registra o actualiza el estado de un contenedor en disco.
func SaveContainer(info ContainerInfo) error {
	if err := os.MkdirAll(defaultStateDir, 0750); err != nil {
		return fmt.Errorf("crear directorio de estado '%s': %w", defaultStateDir, err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar estado de contenedor: %w", err)
	}

	filePath := filepath.Join(defaultStateDir, fmt.Sprintf("%s.json", info.ID))
	return os.WriteFile(filePath, data, 0600)
}

// RemoveContainer elimina el archivo de estado de un contenedor.
func RemoveContainer(id string) error {
	filePath := filepath.Join(defaultStateDir, fmt.Sprintf("%s.json", id))
	return os.Remove(filePath)
}

// IsProcessAlive verifica si un proceso con el PID indicado sigue en ejecución.
func IsProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

// ListContainers retorna la lista de contenedores registrados y actualiza su estado si han terminado.
func ListContainers() ([]ContainerInfo, error) {
	if _, err := os.Stat(defaultStateDir); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := os.ReadDir(defaultStateDir)
	if err != nil {
		return nil, fmt.Errorf("leer directorio de estado: %w", err)
	}

	var containers []ContainerInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(defaultStateDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var info ContainerInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		if info.Status == "running" && !IsProcessAlive(info.PID) {
			info.Status = "stopped"
			_ = SaveContainer(info)
		}

		containers = append(containers, info)
	}

	return containers, nil
}
