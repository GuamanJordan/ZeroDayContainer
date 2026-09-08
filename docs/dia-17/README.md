# Día 17: Inspección y Debug: `setns(2)` y `nsenter`

## Objetivo

Implementar la capacidad de inspeccionar y ejecutar procesos dentro de los namespaces de un contenedor en ejecución (`mc exec`), comprendiendo la llamada al sistema `setns(2)`, la representación de namespaces en `/proc/<pid>/ns/` y el desafío técnico que presenta el modelo multihilo de Go al asociarse a namespaces existentes.

Al finalizar este día debes poder explicar:

- Qué hace la llamada al sistema `setns(2)` y qué descriptores de archivo en `/proc/<pid>/ns/` consume.
- Por qué `setns(CLONE_NEWPID)` y `setns(CLONE_NEWNS)` presentan restricciones estrictas en entornos multihilo (*Go runtime*).
- Cómo `nsenter` resuelve este problema ejecutando la reasociación antes de transferir el control al proceso destino.
- El uso del subcomando `mc exec <pid_or_id> <comando> [args...]` para depurar contenedores en vivo.

---

## La Syscall `setns(2)` y `/proc/<pid>/ns/`

Linux expone los namespaces asignados a cada proceso a través del sistema de archivos pseudo `/proc`:

```text
/proc/<pid>/ns/
├── cgroup -> cgroup:[4026531835]
├── ipc    -> ipc:[4026532210]
├── mnt    -> mnt:[4026532208]
├── net    -> net:[4026532213]
├── pid    -> pid:[4026532211]
├── user   -> user:[4026531837]
└── uts    -> uts:[4026532209]
```

La llamada al sistema:

```c
int setns(int fd, int nstype);
```

permite que el hilo que invoca la llamada se una a los namespaces asociados al descriptor de archivo abierto de `/proc/<pid>/ns/<tipo>`.

### El Reto Multihilo en Go

El runtime de Go utiliza un planificador M:N (múltiples goroutines sobre múltiples hilos de sistema operativo `pthread`). Linux impone que:
1. `setns(fd, CLONE_NEWNS)` falla si el proceso no es monohilo o no puede asegurar atomicidad sobre todos los hilos del proceso.
2. `setns(fd, CLONE_NEWPID)` sólo afecta a los futuros procesos hijos creados por `fork()`, no al proceso invocador actual.

Por estas razones, herramientas como Docker (`runc exec`) y `ZeroDayContainer` utilizan la utilidad estándar `nsenter` (o constructores Cgo antes de inicializar el runtime) para aislar limpiamente la invocación de procesos interactivos o de depuración.

---

## Subcomando `mc exec`

Permite conectarse a un contenedor activo ya sea especificando su **Container ID** registrado o su **PID** directo:

```bash
# Ejecutar un shell dentro del contenedor con ID mc-123456
sudo mc exec mc-123456 /bin/sh

# Ejecutar una inspección directa con PID
sudo mc exec 14209 /bin/ps aux
```

La implementación:
1. Valida que el PID objetivo exista y responda a señales (`kill -0`).
2. Si se proporciona un ID de contenedor, resuelve el PID correspondiente desde `/tmp/mc/containers/`.
3. Invoca `nsenter` configurado con los flags de namespace:
   - `-m`: Mount namespace (acceso al filesystem `rootfs` aislado).
   - `-u`: UTS namespace (mismo hostname).
   - `-i`: IPC namespace (memoria compartida y semáforos).
   - `-n`: Network namespace (mismas interfaces y configuración de red).
   - `-p`: PID namespace (visibilidad de procesos dentro del contenedor).

---

## Checklist de Cierre

- [x] Implementada la función `ExecInContainer` en `internal/namespaces/exec_linux.go` y su stub en `exec_others.go`.
- [x] Pruebas unitarias de validación de argumentos en `internal/namespaces/exec_test.go`.
- [x] Subcomando `mc exec` añadido a `cmd/mc/main.go` con soporte para PID y Container ID.
- [x] Ayuda interactiva de `mc` actualizada con documentación de `mc exec`.
