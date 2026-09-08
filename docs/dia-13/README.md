# Día 13: Hardening de Privilegios — Linux Capabilities y `no_new_privs`

## Objetivo

Comprender y mitigar los riesgos de seguridad derivados de ejecutar contenedores como `root` pleno del sistema host, implementando la reducción del conjunto de **Linux Capabilities** en el proceso contenedor y habilitando la bandera `PR_SET_NO_NEW_PRIVS` mediante la syscall `prctl`.

Al finalizar este día debes poder explicar:

- Qué son las Linux Capabilities y por qué dividen el poder tradicional del superusuario (UID 0).
- La diferencia entre los conjuntos de capabilities: *Bounding Set*, *Permitted*, *Inheritable* y *Effective*.
- Por qué un contenedor nunca debe mantener capabilities como `CAP_SYS_MODULE` o `CAP_SYS_RAWIO`.
- Cómo previene `prctl(PR_SET_NO_NEW_PRIVS, 1)` la escalada de privilegios a través de binarios `setuid` (como `sudo` o `su`).

---

## El Problema del "Root en el Contenedor"

Cuando un proceso inicia en un namespace con UID 0 (root), por defecto hereda casi todas las capabilities del kernel del host. Si el contenedor sufre una brecha, un atacante con `root` pleno dentro del contenedor podría:
- Cargar un módulo malicioso en el kernel del host con `CAP_SYS_MODULE`.
- Modificar la memoria física y hardware del host con `CAP_SYS_RAWIO`.
- Forzar un reinicio del servidor host con `CAP_SYS_BOOT`.
- Alterar la hora global del host con `CAP_SYS_TIME`.

---

## Estrategia de Mitigación: Bounding Set Drop

Linux provee el **Capability Bounding Set**, que actúa como un límite superior inmutable de los privilegios que un proceso (y cualquier hijo que ejecute con `execve`) puede llegar a tener.

Usando la syscall `prctl(PR_CAPBSET_DROP, cap)` eliminamos del bounding set las capabilities críticas:

```go
func DropDangerousCapabilities() error {
	dangerous := []int{
		CAP_SYS_MODULE,    // Bloquea inserción de módulos de kernel
		CAP_SYS_RAWIO,     // Bloquea acceso a /dev/mem y puertos I/O
		CAP_SYS_PACCT,     // Bloquea contabilidad de procesos
		CAP_SYS_ADMIN,     // Bloquea operaciones monolíticas de administración
		CAP_SYS_BOOT,      // Bloquea reboot del host
		CAP_SYS_TIME,      // Bloquea cambio del reloj del sistema
		CAP_AUDIT_CONTROL, // Bloquea manipulación del subsistema de auditoría
		CAP_MAC_OVERRIDE,  // Bloquea anulación de SELinux/AppArmor
		CAP_MAC_ADMIN,     // Bloquea configuración de SELinux/AppArmor
		CAP_WAKE_ALARM,    // Bloquea suspensión/despertar del sistema
		CAP_BLOCK_SUSPEND, // Bloquea bloqueo de suspensión
	}
	for _, cap := range dangerous {
		_ = DropBoundingCapability(cap)
	}
	return nil
}
```

---

## Flag `no_new_privs` (`PR_SET_NO_NEW_PRIVS`)

Invocamos:
```go
prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)
```
Esta directiva del kernel garantiza que ningún subproceso hijo lanzado a partir de ese momento pueda adquirir privilegios adicionales a través de bits `setuid` o `setgid` en el sistema de archivos (bloqueando ataques clásicos de elevación con `/bin/su` o exploits locales).

---

## Verificación en Tiempo de Ejecución

Dentro de un contenedor ejecutado con `mc run`:
```bash
# Ver estado de capabilities del proceso
cat /proc/1/status | grep -i cap
# CapInh: 0000000000000000
# CapPrm: 00000000a80425fb
# CapEff: 00000000a80425fb
# CapBnd: 00000000a80425fb (sin bits de SYS_MODULE, SYS_ADMIN, SYS_BOOT)
```

---

## Checklist de Cierre

- [x] Creado el paquete `internal/capabilities` con `caps_linux.go` y `caps_others.go`.
- [x] Implementadas las funciones `DropDangerousCapabilities` y `EnableNoNewPrivs`.
- [x] Pruebas unitarias en `internal/capabilities/caps_test.go`.
- [x] Integrado `capabilities.RestrictPrivileges()` en el ciclo de vida previo al `execve` en `run_linux.go`.
