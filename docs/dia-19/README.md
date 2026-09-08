# Día 19: Modelo de Amenazas y Postura de Seguridad

## Objetivo

Consolidar la postura de seguridad de `ZeroDayContainer` (`mc`) mediante la elaboración de un **Modelo de Amenazas** formal (basado en la metodología **STRIDE**), el análisis comparativo con otros runtimes líderes de la industria (`runc`, `gVisor`, `Kata Containers`), y la incorporación de análisis estático de vulnerabilidades y seguridad de código en el flujo de integración continua (**CI**).

Al finalizar este día debes poder explicar:

- La aplicación de la metodología STRIDE al ciclo de vida de un contenedor en Linux.
- Las diferencias arquitectónicas fundamentales entre virtualización basada en namespaces (runc, mc), micro-kernels en espacio de usuario (gVisor) y microVMs de hardware (Kata Containers).
- Las capas de defensa en profundidad implementadas en `ZeroDayContainer` (Capabilities pruning, `no_new_privs`, `pivot_root`, Cgroups v2 limits, Network NS).
- Cómo auditar estáticamente el código Go en busca de vulnerabilidades comunes de seguridad.

---

## El Modelo STRIDE en Runtimes de Contenedores

La metodología STRIDE evalúa seis categorías críticas de riesgo en un entorno compartido:

1. **S**poofing (Suplantación de Identidad) -> Mitigado con UTS y Network Namespaces.
2. **T**ampering (Modificación no autorizada) -> Mitigado con `pivot_root`, `MS_RDONLY` y OverlayFS.
3. **R**epudiation (Repudio de acciones) -> Mitigado con persistencia de metadatos en `/tmp/mc/containers/` y métricas de cgroups.
4. **I**nformation Disclosure (Exfiltración o espionaje) -> Mitigado con PID y Mount Namespaces (`/proc` privado).
5. **D**enial of Service (Agotamiento de recursos) -> Mitigado con Cgroups v2 (`memory.max`, `cpu.max`, `pids.max`).
6. **E**levation of Privilege (Escalada de privilegios) -> Mitigado con Bounding Set pruning y `PR_SET_NO_NEW_PRIVS`.

---

## Análisis de Seguridad en CI con `gosec`

Para prevenir vulnerabilidades comunes en tiempo de desarrollo (tales como desbordamientos de enteros, inyecciones de comandos accidentales, o permisos de archivo inseguros), se integra el escáner de seguridad **Gosec** (`securego/gosec`) en GitHub Actions.

El flujo evalúa automáticamente cada `push` y `pull_request` a la rama `main`, garantizando estándares rigurosos de código antes de la fusión.

---

## Checklist de Cierre

- [x] Documentado el modelo formal de amenazas en `docs/THREAT_MODEL.md`.
- [x] Elaborada la matriz comparativa de aislamiento frente a `runc`, `gVisor` y `Kata Containers`.
- [x] Incorporado el job `security-scan` con `gosec` en `.github/workflows/ci.yml`.
- [x] Verificada la coherencia entre las medidas de hardening y la arquitectura del runtime.
