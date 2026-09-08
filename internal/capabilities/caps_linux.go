//go:build linux

package capabilities

import (
	"fmt"
	"syscall"
)

const (
	prSetNoNewPrivs = 38
	prCapbsetRead   = 23
	prCapbsetDrop   = 24
)

// Constantes de Linux Capabilities numéricas
const (
	CAP_SYS_MODULE    = 16
	CAP_SYS_RAWIO     = 17
	CAP_SYS_PACCT     = 20
	CAP_SYS_ADMIN     = 21
	CAP_SYS_BOOT      = 22
	CAP_SYS_TIME      = 25
	CAP_AUDIT_CONTROL = 30
	CAP_MAC_OVERRIDE  = 32
	CAP_MAC_ADMIN     = 33
	CAP_WAKE_ALARM    = 35
	CAP_BLOCK_SUSPEND = 36
)

// GetDefaultDangerousCapabilities retorna la lista de capabilities de alto riesgo que deben ser eliminadas.
func GetDefaultDangerousCapabilities() []int {
	return []int{
		CAP_SYS_MODULE,
		CAP_SYS_RAWIO,
		CAP_SYS_PACCT,
		CAP_SYS_ADMIN,
		CAP_SYS_BOOT,
		CAP_SYS_TIME,
		CAP_AUDIT_CONTROL,
		CAP_MAC_OVERRIDE,
		CAP_MAC_ADMIN,
		CAP_WAKE_ALARM,
		CAP_BLOCK_SUSPEND,
	}
}

// EnableNoNewPrivs invoca prctl(PR_SET_NO_NEW_PRIVS, 1) para evitar que el proceso o sus hijos adquieran
// privilegios adicionales mediante binarios setuid/setgid.
func EnableNoNewPrivs() error {
	_, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, prSetNoNewPrivs, 1, 0, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("prctl(PR_SET_NO_NEW_PRIVS): %w", errno)
	}
	return nil
}

// DropBoundingCapability elimina una capability específica del Bounding Set del proceso actual.
func DropBoundingCapability(cap int) error {
	// Verificar primero si la capability está soportada y activa en el bounding set
	ret, _, _ := syscall.RawSyscall6(syscall.SYS_PRCTL, prCapbsetRead, uintptr(cap), 0, 0, 0, 0)
	if ret != 1 {
		// Ya no está en el bounding set o no está soportada en este kernel
		return nil
	}

	_, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, prCapbsetDrop, uintptr(cap), 0, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("prctl(PR_CAPBSET_DROP, %d): %w", cap, errno)
	}
	return nil
}

// DropDangerousCapabilities remueve las capabilities peligrosas por defecto del Bounding Set.
func DropDangerousCapabilities() error {
	for _, cap := range GetDefaultDangerousCapabilities() {
		if err := DropBoundingCapability(cap); err != nil {
			return err
		}
	}
	return nil
}

// RestrictPrivileges aplica el endurecimiento completo de seguridad antes de ejecutar el workload:
// 1. Elimina capabilities peligrosas del bounding set.
// 2. Activa el flag NO_NEW_PRIVS.
func RestrictPrivileges() error {
	if err := DropDangerousCapabilities(); err != nil {
		return fmt.Errorf("error al reducir capabilities: %w", err)
	}
	if err := EnableNoNewPrivs(); err != nil {
		return fmt.Errorf("error al activar no_new_privs: %w", err)
	}
	return nil
}
