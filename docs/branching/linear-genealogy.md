# Estrategia de ramas: genealogía lineal

Este documento define cómo trabajar las ramas del proyecto para mantener una historia lineal, revisable y sin contaminación entre entregables.

## Objetivo

Cada cambio debe nacer desde `main`, resolver un entregable concreto y volver a `main` solo después de validarse. La rama no debe arrastrar trabajo parcial de otras fases ni mezclar cambios futuros.

## Principios

- `main` representa la única línea estable del proyecto.
- Cada rama temporal resuelve un único entregable.
- Una rama nueva siempre nace del último `main` actualizado.
- No se hacen merges entre ramas temporales.
- Cada PR debe poder revisarse sin depender de código incompleto en otra rama.
- Después de integrar una rama, la siguiente nace desde el nuevo estado de `main`.

## Tipos de ramas

| Prefijo | Uso |
| --- | --- |
| `docs/` | Documentación, notas técnicas, decisiones y guías. |
| `feat/` | Funcionalidad nueva del runtime o CLI. |
| `test/` | Tests, fixtures, CI y validaciones automatizadas. |
| `hardening/` | Seguridad, robustez, limpieza de errores y reducción de riesgo. |
| `release/` | Preparación de versión, empaquetado, changelog y fixes bloqueantes. |

## Flujo de trabajo

```bash
git switch main
git pull --ff-only
git switch -c tipo/NN-descripcion-corta

# implementar solo el entregable de la rama
go build ./...
go test ./...

git push -u origin tipo/NN-descripcion-corta
```

Luego se abre un PR hacia `main`. La integración debe hacerse con squash merge o rebase seguido de fast-forward, según la configuración del repositorio.

Después de integrar:

```bash
git switch main
git pull --ff-only
git branch -d tipo/NN-descripcion-corta
```

## Reglas de contaminación 0

- Una rama no debe modificar archivos fuera del alcance de su entregable.
- Una rama `docs/*` no debe introducir implementación.
- Una rama `test/*` no debe introducir funcionalidad nueva salvo lo mínimo necesario para probar lo ya integrado.
- Una rama `hardening/*` no debe cambiar comportamiento funcional sin test o nota explícita en el PR.
- Una rama `release/*` no debe aceptar features nuevas.
- Si aparece una mejora futura durante el trabajo, se documenta como pendiente y se mueve a otra rama.

## Criterio mínimo para abrir PR

- El objetivo del PR está descrito en una frase.
- El alcance está limitado a un entregable.
- La rama nace desde `main` actualizado.
- Los comandos de validación aplicables fueron ejecutados.
- Los riesgos o limitaciones quedan escritos en la descripción del PR.

## Criterio mínimo para integrar

- CI verde o validación local documentada si la CI todavía no existe.
- Sin cambios mezclados de otra fase.
- Sin archivos generados o temporales innecesarios.
- Sin deuda conocida escondida: cualquier pendiente queda escrito.
