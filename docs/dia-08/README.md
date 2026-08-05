# Día 8: Montajes Esenciales — `/proc`, `/sys`, `/dev` y `/dev/pts`

## Objetivo

Comprender los pseudo-sistemas de archivos esenciales que requieren los contenedores Linux para funcionar correctamente y aprender a montarlos de forma aislada e interactiva dentro del entorno contenedor pivotado.

Al finalizar este día debes poder explicar:

- Para qué sirve y qué información expone cada uno de los 4 montajes esenciales (`/proc`, `/sys`, `/dev`, `/dev/pts`).
- Por qué `/dev` requiere `tmpfs` y nodos/enlaces simbólicos específicos.
- Cómo `MountEssentialFilesystems()` inicializa estos montajes en `mc run`.

---

## Sistemas de Archivos Esenciales

| Punto de Montaje | Tipo (`FSType`) | Descripción / Propósito |
| --- | --- | --- |
| **`/proc`** | `proc` | Expone el estado del kernel y los procesos del PID namespace actual. |
| **`/sys`** | `sysfs` | Expone información de dispositivos del kernel, controladores y cgroups. |
| **`/dev`** | `tmpfs` | Directorio de dispositivos virtuales aislados. |
| **`/dev/pts`** | `devpts` | Instancia aislada para asignación de pseudo-terminales (PTYs). |

---

## Dispositivos Virtuales Básicos en `/dev`

Los binarios Unix (`sh`, `bash`, `cat`, etc.) esperan encontrar ciertos nodos en `/dev` para operaciones básicas de entrada/salida y manejo de descriptores estándar:

- `/dev/null`: Descarta todos los datos escritos.
- `/dev/zero`: Produce un flujo continuo de bytes nulos (`\0`).
- `/dev/random` / `/dev/urandom`: Generadores de números aleatorios del kernel.
- `/dev/tty`: Terminal de control asociada al proceso actual.
- `/dev/stdin` → `/proc/self/fd/0`
- `/dev/stdout` → `/proc/self/fd/1`
- `/dev/stderr` → `/proc/self/fd/2`
- `/dev/ptmx` → `/dev/pts/ptmx`

---

## Implementación en Go (`internal/rootfs/mounts_linux.go`)

```go
func GetEssentialMounts() []MountSpec {
	return []MountSpec{
		{Source: "proc", Target: "/proc", FSType: "proc", Flags: 0, Data: ""},
		{Source: "sysfs", Target: "/sys", FSType: "sysfs", Flags: 0, Data: ""},
		{Source: "tmpfs", Target: "/dev", FSType: "tmpfs", Flags: syscall.MS_NOSUID | syscall.MS_STRICTATIME, Data: "mode=755"},
		{Source: "devpts", Target: "/dev/pts", FSType: "devpts", Flags: 0, Data: "newinstance,ptmxmode=0666,mode=0620,gid=5"},
	}
}
```

---

## Validación Manual dentro de `mc run`

```bash
sudo ./mc run /tmp/alpine-rootfs /bin/sh
```

Verificaciones dentro del contenedor:

1. **Verificar `/proc`:** `ps aux` debe mostrar sólo los procesos del contenedor.
2. **Verificar `/sys`:** `ls /sys` debe listar los subdirectorios `bus`, `class`, `devices`, `fs`, etc.
3. **Verificar `/dev`:** `ls -l /dev` debe mostrar `null`, `zero`, `urandom`, `pts`, `stdin`, `stdout`, `stderr`.
4. **Verificar redirecciones:** `echo "test" > /dev/null` no debe arrojar error.

---

## Checklist de Cierre

- [x] Creada la función `MountEssentialFilesystems` en `internal/rootfs`.
- [x] Configurados los montajes de `proc`, `sysfs`, `tmpfs` y `devpts`.
- [x] Creados enlaces simbólicos para `/dev/fd`, `/dev/stdin`, `/dev/stdout`, `/dev/stderr`.
- [x] Integrados los montajes dentro de `ChildInitPivotRoot`.
- [x] Tests unitarios creados en `internal/rootfs/mounts_test.go`.
