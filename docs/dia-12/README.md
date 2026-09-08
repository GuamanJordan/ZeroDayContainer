# Día 12: Control Groups (Cgroups v2) desde Go

## Objetivo

Implementar la gestión programática de **Cgroups v2** directamente en el runtime de Go a través del paquete `internal/cgroups`, permitiendo limitar el consumo de memoria RAM (`memory.max`), cuota de tiempo de CPU (`cpu.max`) y cantidad máxima de procesos concurrentes (`pids.max`) en el contenedor.

Al finalizar este día debes poder explicar:

- Cómo crear y manipular subdirectorios de cgroups v2 desde Go.
- La conversión de unidades legibles por humanos (ej. `50m`, `1g`) a bytes numéricos exactos para el kernel.
- El cálculo de cuota (`quota`) y periodo (`period`) para el controlador `cpu.max`.
- Cómo asociar el PID de un subproceso hijo recién lanzado al archivo `cgroup.procs` antes de ejecutar el workload.

---

## Arquitectura del Paquete `internal/cgroups`

El paquete `internal/cgroups` encapsula las operaciones de sistema de archivos contra `/sys/fs/cgroup`:

```go
type Config struct {
	MemoryLimit string  // ej: "50m", "100M", "1g"
	CPUs        float64 // ej: 0.5 (50% de 1 core), 1.0, 2.0
	PIDsLimit   int64   // ej: 20
}

type Cgroup struct {
	Name string
	Path string
}
```

### Ciclo de Vida:
1. **Creación:** `cg, err := cgroups.New("zeroday-<id>")` crea `/sys/fs/cgroup/zeroday-<id>`.
2. **Configuración:** `cg.ApplyLimits(cfg)` parsea y escribe en `memory.max`, `cpu.max` y `pids.max`.
3. **Asignación:** `cmd.Start()` inicia el proceso hijo y `cg.AddProcess(cmd.Process.Pid)` lo traslada escribiendo su PID en `cgroup.procs`.
4. **Limpieza:** `defer cg.Cleanup()` elimina el directorio del cgroup al finalizar el contenedor.

---

## Opciones de Límites desde la CLI (`mc run`)

El comando `mc run` soporta las siguientes opciones combinables:

```bash
# Limitar memoria a 50 MB
sudo ./mc run --memory=50m /tmp/alpine-rootfs /bin/sh

# Limitar CPU a medio núcleo (0.5 CPUs)
sudo ./mc run --cpus=0.5 /tmp/alpine-rootfs /bin/sh

# Limitar a máximo 20 procesos concurrentes
sudo ./mc run --pids=20 /tmp/alpine-rootfs /bin/sh

# Combinar límites de recursos y filesystem de solo lectura
sudo ./mc run --read-only --memory=100m --cpus=1.0 --pids=30 /tmp/alpine-rootfs /bin/sh
```

---

## Checklist de Cierre

- [x] Creado el paquete `internal/cgroups` con soporte para cgroups v2 en Linux (`cgroup_linux.go`).
- [x] Implementados stubs de compatibilidad para otros sistemas operativos (`cgroup_others.go`).
- [x] Pruebas unitarias de conversión de memoria y cálculo de cuota de CPU (`cgroup_test.go`).
- [x] Integradas banderas `--memory`, `--cpus` y `--pids` en `cmd/mc/main.go`.
- [x] Asignación y limpieza automática de cgroups por ciclo de vida en `internal/namespaces/run_linux.go`.
