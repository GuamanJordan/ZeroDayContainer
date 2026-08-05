# Día 9: Pruebas de Integración y CI Privilegiado

## Objetivo

Implementar la capa de **pruebas de integración** para verificar el comportamiento real de los namespaces y montajes con permisos de `root`, utilizando la etiqueta de compilación (`build tag`) `//go:build integration`, e integrarla en el pipeline de GitHub Actions.

Al finalizar este día debes poder explicar:

- Por qué los tests de integración deben estar separados de los tests unitarios.
- Cómo usar la directiva `//go:build integration` en Go.
- Cómo ejecutar pruebas que requieren `sudo` localmente y en la CI.
- El rol del job `integration-tests` en GitHub Actions.

---

## Separación: Tests Unitarios vs Tests de Integración

| Característica | Tests Unitarios | Tests de Integración |
| --- | --- | --- |
| **Directiva** | Sin tag especial (se ejecutan por defecto). | `//go:build integration` (solo con `-tags=integration`). |
| **Privilegios** | No requieren permisos de `root`. | Requieren `root` (`sudo` o `euid == 0`). |
| **Propósito** | Probar lógica pura de parsing y constructores. | Probar syscalls reales de Linux (`clone`, `pivot_root`, `mount`). |
| **Entorno** | Cualquier SO / Runner en GitHub Actions. | Linux nativo o runners con soporte Kernel/sudo. |

---

## Ejecución Local de los Tests de Integración

```bash
sudo go test ./test/integration/... -tags=integration -v
```

---

## Configuración en GitHub Actions (`.github/workflows/ci.yml`)

```yaml
  integration-tests:
    name: Integration Tests (Privileged)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run Integration Tests
        run: sudo go test ./test/integration/... -tags=integration -v
```

---

## Checklist de Cierre

- [x] Creado el directorio `test/integration/` y el archivo `integration_test.go`.
- [x] Aplicada la directiva `//go:build integration`.
- [x] Verificada la comprobación de `euid == 0` con `t.Skip` gracioso si falta privilegios.
- [x] Configurado el job `integration-tests` en `.github/workflows/ci.yml`.
- [x] Documentada la estrategia en `docs/dia-09/README.md`.
