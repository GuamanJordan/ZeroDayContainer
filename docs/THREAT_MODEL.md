# Modelo de Amenazas y Arquitectura de Seguridad (Threat Model)

## 1. Resumen Ejecutivo

`ZeroDayContainer` (`mc`) es un runtime de contenedores en Linux desarrollado en Go con fines de investigación, aprendizaje y ejecución modular de procesos aislados. Este documento establece el **Modelo de Amenazas** formal, analizando la superficie de ataque, los límites de confianza, la matriz de aislamiento en comparación con otros estándares de la industria (`runc`, `gVisor`, `Kata Containers`), y las defensas implementadas en el código base.

---

## 2. Límites de Confianza y Supuestos de Seguridad

En el modelo de seguridad de `ZeroDayContainer`, se definen las siguientes fronteras de confianza:

1. **Host y Kernel Linux:** Se asume que el kernel del host no está previamente comprometido y soporta Cgroups v2 y los namespaces modernos de Linux (kernel ≥ 5.4).
2. **Operador del Runtime:** El usuario que ejecuta `mc` cuenta con privilegios administrativos (`root` o `CAP_SYS_ADMIN` en el host).
3. **Carga de Trabajo (Contenedor):** Se considera que el código ejecutado dentro del contenedor es **no confiable** y potencialmente malicioso (código binario arbitrario que intentará escapar, agotar recursos o acceder a datos de otros inquilinos).

---

## 3. Modelo STRIDE Aplicado a Runtimes de Contenedores

| Categoría STRIDE | Vector de Amenaza en Contenedores | Mitigación en `ZeroDayContainer` |
| :--- | :--- | :--- |
| **Spoofing** (Suplantación) | Suplantación de identidad en la red local o spoofing de hostname entre contenedores. | UTS Namespace aislado con hostname propio; Network Namespace privado con par de interfaces `veth` punto a punto. |
| **Tampering** (Alteración) | Modificación indebida del sistema de archivos host o alteración de configuraciones del kernel. | `pivot_root` con destrucción de la raíz antigua (`MNT_DETACH`); opción `--read-only` (`MS_RDONLY`); OverlayFS efímero (`--overlay`); pseudo-fs `/sys` y `/dev` montados con `nodev`, `nosuid` y `noexec`. |
| **Repudiation** (Repudio) | Procesos no rastreables que generan acciones anónimas en el sistema. | Persistencia daemonless en `/tmp/mc/containers/` registrando ID, PID host, comando y timestamp; contabilidad de recursos en Cgroups v2. |
| **Information Disclosure** (Fuga de Información) | Espionaje de procesos vecinos (`/proc`), paquetes de red o memoria IPC compartida. | PID Namespace privado (`CLONE_NEWPID`); `/proc` montado exclusivamente para el nuevo árbol; IPC Namespace privado (`CLONE_NEWIPC`); pila de red aislada (`CLONE_NEWNET`). |
| **Denial of Service** (Denegación de Servicio) | Bombas de bifurcación (*fork bombs*), agotamiento de RAM o consumo del 100% de CPU. | Cgroups v2 jerárquico imponiendo límites estrictos de PIDs (`pids.max`), Memoria (`memory.max`) y cuotas de CPU (`cpu.max`). |
| **Elevation of Privilege** (Escalada de Privilegios) | Abuso de binarios SUID o explotación de Linux Capabilities para adquirir control del host. | Drop del Bounding Set de Capabilities (`CAP_SYS_ADMIN`, `CAP_NET_ADMIN`, etc.); invocación de `prctl(PR_SET_NO_NEW_PRIVS, 1)`. |

---

## 4. Matriz Comparativa de Aislamiento

| Característica / Vector | ZeroDayContainer (`mc`) | runc (Docker / containerd) | gVisor (runsc) | Kata Containers |
| :--- | :--- | :--- | :--- | :--- |
| **Tipo de Aislamiento** | Namespaces + Cgroups v2 + Caps | Namespaces + Cgroups + Caps + Seccomp | Kernel en Userspace (Sentry/Gofer) | MicroVMs hardware (KVM / QEMU) |
| **Kernel Compartido** | Sí (Mismo kernel host) | Sí (Mismo kernel host) | No (Kernel emulado intercepta syscalls) | No (Kernel Linux dedicado por pod/VM) |
| **Filesystem Isolation** | `pivot_root` + OverlayFS | `pivot_root` + OverlayFS | 9P / virtio-fs mediado por Gofer | Imagen de disco en bloque / virtio-fs |
| **Prevención de DoS** | Cgroups v2 nativo | Cgroups v1 / v2 | Límites de proceso + Cgroups | Asignación de vCPU / Memoria por VM |
| **Bounding Set Capabilities** | Sí (Dropping explícito) | Sí (Whitelist OCI) | Sí (Emulado en Sentry) | Sí (Dentro de la VM invitada) |
| **`no_new_privs`** | Sí (`PR_SET_NO_NEW_PRIVS`) | Sí (Activado por defecto) | Sí (Nativo) | Sí (Dentro de la VM invitada) |
| **Filtros Seccomp BPF** | En roadmap (v0.2.0) | Sí (Perfil JSON OCI) | Interceptación total en userspace | Opcional dentro de la VM |
| **Sobrecarga (Overhead)** | Mínima (tiempo de proceso nativo) | Mínima (nativo) | Moderada (intercepción de syscalls) | Ligera/Media (arranque de VM ~100-200ms) |

---

## 5. Medidas de Hardening Implementadas

### 5.1. Restricción de Linux Capabilities y `no_new_privs`
- Antes de ceder el control al binario del usuario mediante `execve`, el contenedor ejecuta `DropDangerousCapabilities()`, removiendo del Bounding Set capacidades críticas como `CAP_SYS_ADMIN`, `CAP_SYS_PTRACE`, `CAP_NET_ADMIN`, `CAP_SYS_RAWIO`, `CAP_MKNOD`, entre otras.
- Se invoca `unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)`. Esto garantiza a nivel de kernel que ningún proceso hijo pueda ganar privilegios adicionales invocando binarios SUID o ejecutables con file capabilities.

### 5.2. Aislamiento del Sistema de Archivos (`pivot_root` y Montajes)
- Se evita el uso de `chroot` en producción por ser susceptible a escapes clásicos (como abrir un FD antes de `chroot` y ejecutar `fchdir(fd)` seguido de múltiples `chroot("..")`).
- `pivot_root` intercambia la raíz del montaje a nivel VFS. La raíz anterior se traslada a un punto temporal (`/oldroot`), se desmonta inmediatamente con `MNT_DETACH` y se elimina el directorio.
- El sistema de archivos host no queda montado en ningún punto accesible dentro del mount namespace del contenedor.

### 5.3. Protección contra DoS con Cgroups v2
- Todos los recursos son gestionados mediante la jerarquía unificada de Cgroups v2.
- `pids.max` neutraliza de manera determinista bombas de procesos (`:(){ :\|:& };:`).
- `memory.max` previene que un contenedor agote la memoria RAM física del host, activando el OOM Killer localizado dentro del cgroup sin desestabilizar procesos fuera de él.
- `cpu.max` impone límites CFS de tiempo de ejecución (cuota/periodo) evitando la saturación de los núcleos del procesador.

### 5.4. Aislamiento de Red
- Cada contenedor con `--net` o `--nat` reside en su propio Network Namespace (`CLONE_NEWNET`), con su propia tabla de enrutamiento, reglas de cortafuegos y loopback aislada.
- La comunicación externa se realiza a través de un par `veth` punto a punto conectado a un bridge administrado (`mc0`), previniendo ataques de ARP spoofing o promiscuous sniffing sobre la interfaz física del host.

---

## 6. Vectores Residuales y Próximos Pasos

Aunque `ZeroDayContainer` implementa las defensas centrales de un contenedor moderno, los siguientes vectores forman parte del roadmap de endurecimiento:

1. **Filtros Seccomp BPF:** Aunque se eliminan capabilities, un exploit de día cero en una syscall del kernel accesible a procesos no privilegiados (e.g. ciertas operaciones `io_uring` o `bpf`) podría comprometer el host. La integración de perfiles Seccomp BPF restringirá las llamadas permitidas a una lista blanca estricta.
2. **User Namespaces (`CLONE_NEWUSER`):** Permitirá ejecutar `ZeroDayContainer` en modo *rootless*, mapeando el UID 0 dentro del contenedor a un UID sin privilegios en el host.
3. **Módulos de Seguridad de Linux (LSM):** Integración futura con perfiles AppArmor y SELinux para control de acceso obligatorio (MAC).
