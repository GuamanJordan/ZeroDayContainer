//go:build linux

package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ExecInContainer entra en los namespaces (mnt, uts, ipc, net, pid) del contenedor
// identificado por targetPID y ejecuta el comando especificado dentro de su contexto.
func ExecInContainer(targetPID int, cmdPath string, args []string) error {
	if targetPID <= 0 {
		return fmt.Errorf("PID inválido: debe ser mayor a 0")
	}
	if cmdPath == "" {
		return fmt.Errorf("se debe especificar un comando para ejecutar")
	}

	// Verificar si el proceso existe enviando la señal 0
	process, err := os.FindProcess(targetPID)
	if err != nil {
		return fmt.Errorf("proceso %d no encontrado: %w", targetPID, err)
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return fmt.Errorf("el contenedor con PID %d no está en ejecución: %w", targetPID, err)
	}

	// Invocar nsenter para reasociar namespaces antes del arranque de hilos
	nsArgs := []string{
		"-t", fmt.Sprintf("%d", targetPID),
		"-m", // Mount namespace
		"-u", // UTS namespace
		"-i", // IPC namespace
		"-n", // Network namespace
		"-p", // PID namespace
		"--",
		cmdPath,
	}
	nsArgs = append(nsArgs, args...)

	cmd := exec.Command("nsenter", nsArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
