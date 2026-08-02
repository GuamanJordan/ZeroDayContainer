# Día 2: Namespaces — Teoría y Exploración Manual

## Objetivo

Comprender la teoría fundamental de los namespaces en el kernel de Linux y aprender a manipularlos de forma manual utilizando herramientas de espacio de usuario (`unshare`, `nsenter`, `ip`, `/proc`).

Al cerrar este día debes poder explicar:

- Qué es un namespace y qué recurso específico aísla cada uno.
- Cómo inspeccionar los namespaces de cualquier proceso mediante `/proc/<PID>/ns/`.
- Cómo crear nuevos namespaces interactivamente con `unshare`.
- Qué ocurre al aislar el PID namespace junto con la re-montadura de `/proc`.
- Cómo interactúan los comandos de espacio de usuario con las syscalls del kernel (`clone`, `unshare`, `setns`).

---

## Modelo mental: ¿Qué es un Namespace?

Un **namespace** en Linux es una abstracción que envuelve un recurso global del sistema para que los procesos dentro del namespace tengan la ilusión de poseer su propia instancia aislada de dicho recurso.

Sin namespaces, todos los procesos comparten la misma vista global del hostname, las interfaces de red, la tabla de procesos, los puntos de montaje, etc.

---

## Los 7+1 Namespaces de Linux

| Namespace | Flag de `clone` / `unshare` | Recurso Aislado |
| --- | --- | --- |
| **UTS** (UNIX Timesharing System) | `CLONE_NEWUTS` | Hostname y nombre de dominio NIS. |
| **PID** (Process ID) | `CLONE_NEWPID` | Árbol de PIDs (el primer proceso es PID 1 dentro del namespace). |
| **Mount** (MNT) | `CLONE_NEWNS` | Puntos de montaje del sistema de archivos (`/`, `/proc`, `/sys`, etc.). |
| **Network** (NET) | `CLONE_NEWNET` | Interfaces de red, tablas de ruteo, sockets, puertos firewall (iptables/nftables). |
| **IPC** (Inter-Process Comm) | `CLONE_NEWIPC` | Recurso de comunicación entre procesos (System V IPC, colas de mensajes POSIX). |
| **User** | `CLONE_NEWUSER` | UIDs y GIDs (permite ser root UID 0 dentro del namespace sin ser root en el host). |
| **Cgroup** | `CLONE_NEWCGROUP` | Vista de la jerarquía de cgroups (`/proc/self/cgroup`). |
| **Time** (kernel ≥ 5.6) | `CLONE_NEWTIME` | Reloj del sistema (`CLOCK_MONOTONIC`, `CLOCK_REALTIME`). |

---

## Exploración Manual Paso a Paso

### 1. Inspección de Namespaces en `/proc`

Cada proceso en Linux expone sus namespaces asignados en el pseudodirectorio `/proc/<PID>/ns/`. Cada archivo es un enlace simbólico especial cuyo inode identifica el namespace concreto.

```bash
ls -l /proc/$$/ns/
```

**Salida típica:**

```text
lrwxrwxrwx 1 user user 0 Aug  1 23:15 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 mnt -> 'mnt:[4026531840]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 net -> 'net:[4026531992]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 pid_for_children -> 'pid:[4026531836]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 user -> 'user:[4026531837]'
lrwxrwxrwx 1 user user 0 Aug  1 23:15 uts -> 'uts:[4026531838]'
```

Si dos procesos comparten el mismo número de inode para un namespace (por ejemplo `uts:[4026531838]`), significa que comparten ese recurso. Si el inode cambia, están aislados.

---

### 2. Experimento A: Aislamiento de UTS Namespace (Hostname)

El UTS namespace permite cambiar el nombre del host dentro de la shell aislada sin afectar al host principal.

**Paso 1:** Abrir un nuevo UTS namespace con `unshare`:
```bash
sudo unshare --uts bash
```

**Paso 2:** Cambiar el hostname dentro de la nueva shell:
```bash
hostname contenedor-demo
hostname
```

**Paso 3:** En otra terminal del host principal, verificar el hostname:
```bash
hostname
```

**Resultado:** En la shell aislada el hostname es `contenedor-demo`, mientras que en el host principal se mantiene intacto.

---

### 3. Experimento B: Aislamiento de PID + Mount Namespace

Por defecto, si solo aíslas el PID namespace con `unshare --pid`, el comando `ps` seguirá mostrando todos los procesos del host porque `ps` lee el sistema de archivos `/proc` que aún apunta al del host.

Para lograr un aislamiento completo del árbol de procesos se requiere combinar `--pid`, `--fork` y `--mount-proc`:

```bash
sudo unshare --pid --fork --mount-proc bash
```

**Verificaciones dentro de la shell aislada:**

1. Verificar el PID del proceso actual:
   ```bash
   echo $$
   # Salida: 1
   ```
2. Verificar el árbol de procesos:
   ```bash
   ps aux
   ```

**Salida observada:**
```text
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root           1  0.0  0.1   8956  5120 pts/0    S    23:20   0:00 bash
root           8  0.0  0.0  10616  3328 pts/0    R+   23:20   0:00 ps aux
```

El proceso `bash` se ha convertido en **PID 1** dentro de su propio árbol de procesos.

---

### 4. Experimento C: Aislamiento de Network Namespace

El Network namespace aísla las interfaces de red, tablas de ruteo y sockets.

**Crear un network namespace llamado `demo-net`:**

```bash
sudo ip netns add demo-net
```

**Listar interfaces de red dentro del namespace:**

```bash
sudo ip netns exec demo-net ip a
```

**Salida observada:**
```text
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
```

Se observa que sólo existe la interfaz `lo` (loopback) desactivada. El namespace de red inicia totalmente aislado sin acceso a las placas de red físicas ni a internet.

**Limpieza:**
```bash
sudo ip netns del demo-net
```

---

## Syscalls Fundamentales detrás de los Namespaces

1. **`clone(..., flags)`**: Crea un nuevo proceso hijo permitiendo especificar los flags `CLONE_NEW*` para crearlo en nuevos namespaces.
2. **`unshare(flags)`**: Mueve el proceso actual (o crea uno nuevo) hacia un nuevo namespace sin necesidad de forkar.
3. **`setns(fd, nstype)`**: Permite a un proceso adjuntarse (unirse) a un namespace existente a través de su descriptor de archivo (por ejemplo, `/proc/<PID>/ns/net`). Es la base del comando `docker exec` / `mc exec`.

---

## Relación con el Runtime `mc`

En Go, el runtime `mc` que estamos construyendo utiliza estas mismas syscalls a través de la librería estándar `os/exec` y el paquete `syscall`:

```go
cmd := exec.Command("/proc/self/exe", append([]string{"child-init", cmdPath}, args...)...)
cmd.SysProcAttr = &syscall.SysProcAttr{
	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
}
```

- `CLONE_NEWUTS`: Equivale a `unshare --uts`.
- `CLONE_NEWPID`: Equivale a `unshare --pid`.
- `CLONE_NEWNS`: Equivale a `unshare --mount`.

---

## Checklist de Cierre

- [x] Entendí qué aísla cada uno de los 7+1 namespaces en Linux.
- [x] Inspeccioné `/proc/$$/ns/` y verifiqué los inodes de los namespaces.
- [x] Ejecuté la prueba manual de UTS con `unshare --uts`.
- [x] Ejecuté la prueba de PID y `/proc` privado con `unshare --pid --fork --mount-proc`.
- [x] Ejecuté la prueba de Network namespace con `ip netns`.
- [x] Puedo explicar la diferencia entre `clone`, `unshare` y `setns`.

---

## Validación Sugerida

```bash
go build ./...
go test ./...
git diff --check
```
