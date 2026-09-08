package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/GuamanJordan/ZeroDayContainer/internal/cgroups"
	"github.com/GuamanJordan/ZeroDayContainer/internal/namespaces"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run-basic":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		if err := namespaces.RunBasic(os.Args[2], os.Args[3:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "run-chroot":
		if len(os.Args) < 4 {
			printUsage()
			os.Exit(1)
		}
		if err := namespaces.RunChroot(os.Args[2], os.Args[3], os.Args[4:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "run", "run-pivot":
		readOnly := false
		var cgCfg cgroups.Config
		i := 2
		for i < len(os.Args) {
			arg := os.Args[i]
			if arg == "--read-only" || arg == "-ro" {
				readOnly = true
				i++
			} else if strings.HasPrefix(arg, "--memory=") {
				cgCfg.MemoryLimit = strings.TrimPrefix(arg, "--memory=")
				i++
			} else if strings.HasPrefix(arg, "--cpus=") {
				cpuStr := strings.TrimPrefix(arg, "--cpus=")
				if val, err := strconv.ParseFloat(cpuStr, 64); err == nil {
					cgCfg.CPUs = val
				}
				i++
			} else if strings.HasPrefix(arg, "--pids=") {
				pidsStr := strings.TrimPrefix(arg, "--pids=")
				if val, err := strconv.ParseInt(pidsStr, 10, 64); err == nil {
					cgCfg.PIDsLimit = val
				}
				i++
			} else {
				break
			}
		}
		if len(os.Args) < i+2 {
			printUsage()
			os.Exit(1)
		}
		rootfsPath := os.Args[i]
		cmdPath := os.Args[i+1]
		cmdArgs := os.Args[i+2:]
		if err := namespaces.RunPivotRootWithCgroups(rootfsPath, readOnly, cgCfg, cmdPath, cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "child-init":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		if err := namespaces.ChildInit(os.Args[2], os.Args[3:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "child-init-chroot":
		if len(os.Args) < 4 {
			printUsage()
			os.Exit(1)
		}
		if err := namespaces.ChildInitChroot(os.Args[2], os.Args[3], os.Args[4:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "child-init-pivot":
		readOnly := false
		argsStart := 2
		if len(os.Args) > 2 && (os.Args[2] == "--read-only" || os.Args[2] == "-ro") {
			readOnly = true
			argsStart = 3
		}
		if len(os.Args) < argsStart+2 {
			printUsage()
			os.Exit(1)
		}
		rootfsPath := os.Args[argsStart]
		cmdPath := os.Args[argsStart+1]
		cmdArgs := os.Args[argsStart+2:]
		if err := namespaces.ChildInitPivotRoot(rootfsPath, readOnly, cmdPath, cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("comando desconocido:", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("uso: mc <comando> [argumentos]")
	fmt.Println("comandos disponibles:")
	fmt.Println("  run [opciones] <rootfs_path> <comando> [args...] Lanza un contenedor completo aislado con pivot_root")
	fmt.Println("    Opciones:")
	fmt.Println("      --read-only, -ro           Monta el rootfs en modo solo lectura")
	fmt.Println("      --memory=<limite>          Límite de memoria (ej: 50m, 1g)")
	fmt.Println("      --cpus=<cores>             Límite de CPU cores (ej: 0.5, 1.0)")
	fmt.Println("      --pids=<max>               Límite de procesos concurrentes (ej: 50)")
	fmt.Println("  run-basic <comando> [args...]                     Lanza un subproceso aislado en UTS, PID y Mount namespaces")
	fmt.Println("  run-chroot <rootfs_path> <comando> [args...]        Lanza un proceso aislado con chroot")
}
