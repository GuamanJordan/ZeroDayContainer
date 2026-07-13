# ADR 0001: Modelo de ramas con genealogía lineal

## Estado

Aceptado.

## Contexto

El proyecto se desarrollará por entregables diarios durante un mes. Cada día introduce conceptos técnicos con dependencias claras: namespaces, filesystem, cgroups, red, CLI, seguridad y release.

Si se permiten ramas largas, merges entre features o trabajo parcial acumulado, el historial se vuelve difícil de auditar. También aumenta el riesgo de mezclar implementación futura con entregables anteriores.

El proyecto necesita una forma simple de mantener trazabilidad y revisión limpia.

## Decisión

Se usará un modelo de ramas temporales con genealogía lineal:

- `main` será la única rama permanente y estable.
- Cada entregable tendrá una rama temporal.
- Cada rama nacerá del último `main` actualizado.
- No se harán merges entre ramas temporales.
- Cada rama se integrará a `main` solo después de validarse.
- La historia final de `main` deberá quedar lineal.

Los prefijos permitidos son:

- `docs/`
- `feat/`
- `test/`
- `hardening/`
- `release/`

## Consecuencias

### Positivas

- La revisión de PRs es más simple.
- Cada commit integrado corresponde a un entregable claro.
- Se reduce la contaminación entre fases.
- Es más fácil revertir un cambio específico.
- El cronograma del plan se puede mapear directamente a ramas.

### Negativas

- Puede haber más PRs pequeños.
- Algunas tareas grandes deben dividirse con disciplina.
- El trabajo parcial no integrado debe documentarse en vez de quedarse oculto en una rama larga.

## Alternativas consideradas

### Ramas largas por semana

Rechazada porque mezcla varios entregables y hace más difícil detectar qué cambio introdujo una regresión.

### Git flow completo

Rechazado porque agrega ramas permanentes y ceremonias innecesarias para un proyecto educativo de un mes.

### Trunk based development sin ramas

Rechazado porque el proyecto necesita PRs revisables y separación explícita entre entregables diarios.

## Criterio de éxito

El modelo funciona si `main` mantiene una historia lineal, cada PR se revisa de forma aislada y ningún entregable depende de código incompleto en otra rama.
