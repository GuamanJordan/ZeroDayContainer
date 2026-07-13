# Checklist de Pull Request

Usa este checklist antes de abrir o integrar cualquier PR.

## Identidad del PR

- [ ] El título sigue el formato `tipo: descripcion corta`.
- [ ] La rama usa el formato `tipo/NN-descripcion-corta`.
- [ ] El PR apunta a `main`.
- [ ] La rama nació desde `main` actualizado.

## Alcance

- [ ] El PR resuelve un solo entregable.
- [ ] No mezcla documentación, feature, tests, hardening o release sin justificación.
- [ ] No incluye trabajo parcial de fases futuras.
- [ ] No modifica archivos fuera del alcance declarado.

## Descripción sugerida

```md
## Objetivo

Describe en una frase qué entrega este PR.

## Cambios

- Cambio principal 1.
- Cambio principal 2.

## Validación

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `sudo go test ./test/integration/... -tags=integration -v`

## Riesgos

- Riesgo o limitación conocida.

## Fuera de alcance

- Trabajo que queda para otra rama.
```

## Validación por tipo de rama

| Tipo | Validación mínima |
| --- | --- |
| `docs/*` | Revisión de enlaces, nombres de archivos y coherencia con el plan. |
| `feat/*` | `go build ./...` y `go test ./...`. |
| `test/*` | Tests nuevos fallan antes del cambio si aplica y pasan después. |
| `hardening/*` | Tests aplicables, nota de riesgo y comportamiento esperado. |
| `release/*` | Pipeline completo verde antes de crear tag. |

## Antes de integrar

- [ ] El diff es pequeño y revisable.
- [ ] La CI está verde o la excepción está documentada.
- [ ] No hay conflictos con `main`.
- [ ] No hay secretos, archivos temporales ni binarios accidentales.
- [ ] Los pendientes están escritos como issues, TODOs documentales o ramas futuras.

## Después de integrar

- [ ] Actualizar `main` local con `git pull --ff-only`.
- [ ] Borrar la rama local.
- [ ] Borrar la rama remota si el hosting no lo hizo automáticamente.
- [ ] Crear la siguiente rama desde el nuevo `main`.
