//go:build linux

package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const cgroupRoot = "/sys/fs/cgroup"

// Config define las restricciones de recursos para un cgroup v2.
type Config struct {
	MemoryLimit string  // ej: "50m", "100M", "1g"
	CPUs        float64 // ej: 0.5 (50% de 1 núcleo), 1.0, 2.0
	PIDsLimit   int64   // ej: 20
}

// Cgroup representa un cgroup v2 creado en /sys/fs/cgroup.
type Cgroup struct {
	Name string
	Path string
}

// ParseMemoryBytes convierte cadenas como "50m", "100M", "1g", "2G" a bytes.
func ParseMemoryBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, nil
	}

	unitMultiplier := int64(1)
	valStr := s

	switch {
	case strings.HasSuffix(s, "kb"):
		unitMultiplier = 1024
		valStr = strings.TrimSuffix(s, "kb")
	case strings.HasSuffix(s, "k"):
		unitMultiplier = 1024
		valStr = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "mb"):
		unitMultiplier = 1024 * 1024
		valStr = strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "m"):
		unitMultiplier = 1024 * 1024
		valStr = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "gb"):
		unitMultiplier = 1024 * 1024 * 1024
		valStr = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "g"):
		unitMultiplier = 1024 * 1024 * 1024
		valStr = strings.TrimSuffix(s, "g")
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("formato de memoria no válido '%s': %w", s, err)
	}

	return val * unitMultiplier, nil
}

// FormatCPUQuotaPeriod calcula el quota y period para cpu.max a partir de un float de cores.
// Ejemplo: cpus = 0.5 -> quota 50000, period 100000.
func FormatCPUQuotaPeriod(cpus float64) (string, error) {
	if cpus <= 0 {
		return "", fmt.Errorf("el límite de CPU debe ser mayor a 0")
	}

	const period = 100000 // 100ms en microsegundos
	quota := int64(cpus * float64(period))
	return fmt.Sprintf("%d %d", quota, period), nil
}

// New crea un nuevo cgroup v2 bajo /sys/fs/cgroup/<name>.
func New(name string) (*Cgroup, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre del cgroup no puede estar vacío")
	}

	cgPath := filepath.Join(cgroupRoot, name)
	if err := os.MkdirAll(cgPath, 0750); err != nil {
		return nil, fmt.Errorf("crear cgroup en '%s': %w", cgPath, err)
	}

	return &Cgroup{
		Name: name,
		Path: cgPath,
	}, nil
}

// ApplyLimits escribe las directivas en los archivos del cgroup según la configuración.
func (cg *Cgroup) ApplyLimits(cfg Config) error {
	// 1. Límite de memoria (memory.max)
	if cfg.MemoryLimit != "" {
		memBytes, err := ParseMemoryBytes(cfg.MemoryLimit)
		if err != nil {
			return err
		}
		if memBytes > 0 {
			memFile := filepath.Join(cg.Path, "memory.max")
			if err := os.WriteFile(memFile, []byte(strconv.FormatInt(memBytes, 10)), 0600); err != nil {
				return fmt.Errorf("escribir memory.max: %w", err)
			}
		}
	}

	// 2. Límite de CPU (cpu.max)
	if cfg.CPUs > 0 {
		cpuVal, err := FormatCPUQuotaPeriod(cfg.CPUs)
		if err != nil {
			return err
		}
		cpuFile := filepath.Join(cg.Path, "cpu.max")
		if err := os.WriteFile(cpuFile, []byte(cpuVal), 0600); err != nil {
			return fmt.Errorf("escribir cpu.max: %w", err)
		}
	}

	// 3. Límite de PIDs (pids.max)
	if cfg.PIDsLimit > 0 {
		pidsFile := filepath.Join(cg.Path, "pids.max")
		if err := os.WriteFile(pidsFile, []byte(strconv.FormatInt(cfg.PIDsLimit, 10)), 0600); err != nil {
			return fmt.Errorf("escribir pids.max: %w", err)
		}
	}

	return nil
}

// AddProcess añade el PID indicado al archivo cgroup.procs.
func (cg *Cgroup) AddProcess(pid int) error {
	procsFile := filepath.Join(cg.Path, "cgroup.procs")
	if err := os.WriteFile(procsFile, []byte(strconv.Itoa(pid)), 0600); err != nil {
		return fmt.Errorf("añadir pid %d a cgroup.procs: %w", pid, err)
	}
	return nil
}

// Cleanup elimina el directorio del cgroup.
func (cg *Cgroup) Cleanup() error {
	if cg.Path == "" || cg.Path == cgroupRoot {
		return nil
	}
	return os.Remove(cg.Path)
}
