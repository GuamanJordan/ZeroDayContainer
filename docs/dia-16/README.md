# Día 16: Diseño y Arquitectura de la CLI (`mc`)

## Objetivo

Estructurar una interfaz de línea de comandos (**CLI**) moderna, intuitiva y modular para `ZeroDayContainer` (`mc`), implementando persistencia y consulta de estado de contenedores sin requerir un demonio central en segundo plano (arquitectura *daemonless*, similar a `podman` o `crun`).

Al finalizar este día debes poder explicar:

- Las diferencias arquitectónicas entre un runtime dependiente de un demonio (como Docker Engine) y un runtime *daemonless* (como Podman o mc).
- Cómo rastrear procesos y metadatos de contenedores en `/tmp/mc/containers/` en archivos JSON ligeros.
- La técnica de sondeo `process.Signal(syscall.Signal(0))` para detectar procesos vivos en Linux sin afectarlos.
- La organización de subcomandos (`run`, `list`/`ps`, `version`, `run-basic`).

---

## Arquitectura *Daemonless*: Sin Demonio Central

A diferencia de Docker (donde un proceso demonio `dockerd` corre continuamente como root gestionando sockets gRPC/REST), `mc` utiliza el modelo de proceso directo:
- Cada comando `mc run` crea su árbol de namespaces y cgroups directamente.
- El estado de los contenedores se registra en el sistema de archivos (`/tmp/mc/containers/<id>.json`).
- Si el proceso contenedor muere, la comprobación `kill(PID, 0)` detecta que el proceso ya no existe y actualiza el estado a `stopped`.

```text
mc run  ----(escribe JSON)----> [ /tmp/mc/containers/<id>.json ]
                                              ^
mc list <---(lee y verifica PID 0)------------+
```

---

## Subcomandos Disponibles en `mc`

| Subcomando | Propósito | Ejemplo de Uso |
| --- | --- | --- |
| **`mc run`** | Ejecuta un contenedor aislado con `pivot_root`, cgroups y red opcional. | `sudo mc run --net --memory=100m /tmp/rootfs /bin/sh` |
| **`mc list` / `mc ps`** | Muestra la tabla de contenedores registrados y su estado. | `mc list` |
| **`mc version`** | Imprime la versión del runtime y detalles de arquitectura. | `mc version` |
| **`mc run-basic`** | Ejecuta un subproceso mínimo aislado (UTS, PID, Mount). | `sudo mc run-basic /bin/bash` |

---

## Formato de Salida de `mc list`

```text
CONTAINER ID    PID      STATUS     COMMAND              CREATED
mc-83921        14209    running    /bin/sh              2026-09-08 11:30:15
mc-90144        14350    stopped    /bin/busybox         2026-09-08 11:28:02
```

---

## Checklist de Cierre

- [x] Creado el paquete `internal/state` con persistencia en JSON y detección de vida vía `Signal(0)`.
- [x] Pruebas unitarias de serialización y ciclo de vida en `internal/state/state_test.go`.
- [x] Agregados los subcomandos `list`/`ps` y `version` en `cmd/mc/main.go`.
- [x] Registro automático y desregistro al terminar en `internal/namespaces/run_linux.go`.
