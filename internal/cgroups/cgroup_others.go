//go:build !linux

package cgroups

import (
	"fmt"
)

// Config define las restricciones de recursos para un cgroup.
type Config struct {
	MemoryLimit string
	CPUs        float64
	PIDsLimit   int64
}

// Cgroup representa un cgroup.
type Cgroup struct {
	Name string
	Path string
}

// ParseMemoryBytes en plataformas no-Linux convierte memoria a bytes.
func ParseMemoryBytes(s string) (int64, error) {
	return 0, nil
}

// FormatCPUQuotaPeriod en plataformas no-Linux calcula quota y period.
func FormatCPUQuotaPeriod(cpus float64) (string, error) {
	return "", nil
}

// New en plataformas no-Linux retorna un error de incompatibilidad.
func New(name string) (*Cgroup, error) {
	return nil, fmt.Errorf("cgroups v2 solo está soportado en Linux")
}

// ApplyLimits en plataformas no-Linux retorna un error.
func (cg *Cgroup) ApplyLimits(cfg Config) error {
	return fmt.Errorf("cgroups v2 solo está soportado en Linux")
}

// AddProcess en plataformas no-Linux retorna un error.
func (cg *Cgroup) AddProcess(pid int) error {
	return fmt.Errorf("cgroups v2 solo está soportado en Linux")
}

// Cleanup en plataformas no-Linux es un no-op.
func (cg *Cgroup) Cleanup() error {
	return nil
}
