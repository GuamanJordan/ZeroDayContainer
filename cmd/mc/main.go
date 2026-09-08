package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/GuamanJordan/ZeroDayContainer/internal/cgroups"
	"github.com/GuamanJordan/ZeroDayContainer/internal/namespaces"
	"github.com/GuamanJordan/ZeroDayContainer/internal/state"
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
		enableNet := false
		enableNAT := false
		enableOverlay := false
		var cgCfg cgroups.Config
		i := 2
		for i < len(os.Args) {
			arg := os.Args[i]
			if arg == "--read-only" || arg == "-ro" {
				readOnly = true
				i++
			} else if arg == "--net" {
				enableNet = true
				i++
			} else if arg == "--nat" {
				enableNAT = true
				enableNet = true
				i++
			} else if arg == "--overlay" {
				enableOverlay = true
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
		opts := namespaces.ContainerOpts{
			ReadOnly:      readOnly,
			EnableNet:     enableNet,
			EnableNAT:     enableNAT,
			EnableOverlay: enableOverlay,
			Cgroups:       cgCfg,
		}
		if err := namespaces.RunPivotRootWithOptions(rootfsPath, opts, cmdPath, cmdArgs); err != nil {
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
	case "list", "ps":
		containers, err := state.ListContainers()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error al listar contenedores:", err)
			os.Exit(1)
		}
		if len(containers) == 0 {
			fmt.Println("No hay contenedores registrados en ejecución.")
			return
		}
		fmt.Printf("%-15s %-8s %-10s %-20s %s\n", "CONTAINER ID", "PID", "STATUS", "COMMAND", "CREATED")
		for _, c := range containers {
			createdStr := c.CreatedAt.Format("2006-01-02 15:04:05")
			fmt.Printf("%-15s %-8d %-10s %-20s %s\n", c.ID, c.PID, c.Status, c.Command, createdStr)
		}
	case "exec":
		if len(os.Args) < 4 {
			fmt.Println("uso: mc exec <pid_or_id> <comando> [args...]")
			os.Exit(1)
		}
		target := os.Args[2]
		cmdPath := os.Args[3]
		cmdArgs := os.Args[4:]

		targetPID, err := strconv.Atoi(target)
		if err != nil {
			containers, listErr := state.ListContainers()
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error al buscar contenedor: %v\n", listErr)
				os.Exit(1)
			}
			found := false
			for _, c := range containers {
				if c.ID == target {
					targetPID = c.PID
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "no se encontró ningún contenedor activo con ID o PID %q\n", target)
				os.Exit(1)
			}
		}

		if err := namespaces.ExecInContainer(targetPID, cmdPath, cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "version", "-v", "--version":
		fmt.Println("ZeroDayContainer (mc) v0.1.0")
		fmt.Printf("Go runtime: %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		fmt.Println("Runtime modular de contenedores en Linux sin Docker ni runc")
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
	fmt.Println("      --overlay                  Monta una capa OverlayFS efímera sobre el rootfs base")
	fmt.Println("      --net                      Aísla la red con network namespace y par de interfaces veth")
	fmt.Println("      --nat                      Conecta al bridge mc0 y habilita NAT para salida a internet")
	fmt.Println("      --memory=<limite>          Límite de memoria (ej: 50m, 1g)")
	fmt.Println("      --cpus=<cores>             Límite de CPU cores (ej: 0.5, 1.0)")
	fmt.Println("      --pids=<max>               Límite de procesos concurrentes (ej: 50)")
	fmt.Println("  exec <pid_or_id> <comando> [args...]              Ejecuta un comando en los namespaces de un contenedor activo")
	fmt.Println("  list, ps                                          Lista los contenedores registrados")
	fmt.Println("  version                                           Muestra la versión del runtime")
	fmt.Println("  run-basic <comando> [args...]                     Lanza un subproceso aislado en UTS, PID y Mount namespaces")
	fmt.Println("  run-chroot <rootfs_path> <comando> [args...]        Lanza un proceso aislado con chroot")
}

