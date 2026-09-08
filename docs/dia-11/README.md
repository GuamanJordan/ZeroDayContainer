# Día 11: Control Groups (Cgroups v2) — Teoría y Exploración Manual

## Objetivo

Comprender la arquitectura y funcionamiento de **Cgroups v2** (Control Groups versión 2) en Linux, aprender cómo el kernel administra la jerarquía unificada de recursos (memoria, CPU, procesos) y explorar manualmente la creación y configuración de límites mediante el pseudo-sistema de archivos `/sys/fs/cgroup`.

Al finalizar este día debes poder explicar:

- Qué son los cgroups y en qué se diferencian conceptualmente de los namespaces.
- Por qué Linux migró de la jerarquía múltiple de cgroups v1 a la jerarquía unificada de cgroups v2.
- Cómo se aplican límites de recursos sin necesidad de syscalls especiales, interactuando únicamente con archivos de texto.
- Cómo el kernel ejecuta el **OOM Killer** cuando un proceso excede `memory.max`.

---

## Namespaces vs Cgroups: Las dos caras de un contenedor

En los contenedores Linux existe una distinción fundamental entre el aislamiento de visión y la restricción de consumo:

| Dimensión | Namespaces (Espacios de Nombres) | Cgroups (Grupos de Control) |
| --- | --- | --- |
| **Pregunta que responde** | *"¿Qué puede **ver** el proceso?"* | *"¿Cuánto puede **consumir** el proceso?"* |
| **Recursos afectados** | PIDs, interfaces de red, puntos de montaje, hostname, IPC, usuarios. | Memoria RAM, ciclos de CPU, número de procesos (PIDs), I/O de disco. |
| **Mecanismo del Kernel** | Flags en `clone()` (`CLONE_NEWPID`, etc.), syscalls `unshare`, `setns`. | Jerarquía de directorios y archivos de control en `/sys/fs/cgroup/`. |
| **Consecuencia de fallo** | El proceso no ve recursos del host ni de otros contenedores. | Si se pasa del límite, el kernel estrangula la CPU o mata el proceso (OOM Kill). |

---

## Cgroups v1 vs Cgroups v2

En **cgroups v1**, cada controlador (memory, cpu, blkio, cpuset) tenía su propia jerarquía independiente en `/sys/fs/cgroup/memory`, `/sys/fs/cgroup/cpu`, lo que causaba inconsistencias complejas al coordinar límites entre memoria y I/O de páginas sucias.

En **cgroups v2** (estándar desde Linux 4.5 y por defecto en distribuciones modernas):
1. **Jerarquía Unificada:** Existe un único árbol en `/sys/fs/cgroup`.
2. **Regla "No internal processes":** Un proceso solo puede residir en nodos hoja si el grupo tiene controladores hijos activados.
3. **Controladores habilitados por delegación:** Se activan agregando `+memory +cpu +pids` en `cgroup.subtree_control`.

### Verificar soporte de Cgroups v2 en tu sistema
```bash
cat /sys/fs/cgroup/cgroup.controllers
# Salida esperada (ejemplo):
# cpuset cpu io memory hugetlb pids rdma misc
```

---

## Controladores Clave en Cgroups v2

1. **`memory.max`**: Límite estricto de memoria RAM en bytes. Si los procesos del cgroup superan este límite y no pueden liberar memoria mediante reclaim, el OOM killer termina el proceso más consumidor dentro del grupo.
2. **`cpu.max`**: Asignación de cuota y periodo de tiempo de CPU (`$QUOTA $PERIOD`). Ejemplo: `50000 100000` equivale a 50% de 1 núcleo (0.5 CPUs).
3. **`pids.max`**: Número máximo de procesos/hilos permitidos dentro del cgroup (protección contra *fork bombs*).
4. **`cgroup.procs`**: Archivo donde se escribe el PID de un proceso para moverlo al cgroup (el kernel mueve automáticamente todos sus hilos asociados).
5. **`memory.events`**: Monitoreo de eventos como `low`, `high`, `max`, `oom`, `oom_kill`.

---

## Práctica Manual: Creación de un Cgroup y Prueba del OOM Killer

### Paso 1: Crear un cgroup de prueba
```bash
sudo mkdir /sys/fs/cgroup/mc-test
```
El kernel de Linux poblará automáticamente este directorio con todos los archivos de control asociados.

### Paso 2: Establecer límite de memoria a 50 MB
```bash
echo "52428800" | sudo tee /sys/fs/cgroup/mc-test/memory.max
```

### Paso 3: Asignar una shell al cgroup
En una terminal:
```bash
echo $$ | sudo tee /sys/fs/cgroup/mc-test/cgroup.procs
```
Verifica que el proceso se haya movido:
```bash
cat /sys/fs/cgroup/mc-test/cgroup.procs
```

### Paso 4: Provocar un Out Of Memory (OOM)
Ejecutar un comando o script que intente reservar más de 50 MB en memoria (por ejemplo con python3):
```bash
python3 -c "a = ' ' * (60 * 1024 * 1024)"
# Salida esperada:
# Killed (o Terminado por señal SIGKILL)
```

### Paso 5: Inspeccionar los eventos de OOM
```bash
cat /sys/fs/cgroup/mc-test/memory.events
# Salida:
# low 0
# high 0
# max 1
# oom 1
# oom_kill 1
```

### Paso 6: Limpieza
Mover la shell de regreso al cgroup raíz y eliminar el grupo:
```bash
echo $$ | sudo tee /sys/fs/cgroup/cgroup.procs
sudo rmdir /sys/fs/cgroup/mc-test
```

---

## Checklist de Cierre

- [x] Verificado el montaje de cgroups v2 en el sistema host.
- [x] Documentadas las diferencias entre Namespaces y Cgroups.
- [x] Explicada la estructura de archivos en cgroups v2 (`memory.max`, `cpu.max`, `pids.max`, `cgroup.procs`).
- [x] Realizada la prueba práctica manual de OOM Killer con `memory.events`.
