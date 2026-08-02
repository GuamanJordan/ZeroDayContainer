package main

import (
	"fmt"
	"os"

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
	case "child-init":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		if err := namespaces.ChildInit(os.Args[2], os.Args[3:]); err != nil {
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
	fmt.Println("  run-basic <comando> [args...]   Lanza un subproceso aislado en UTS, PID y Mount namespaces")
}
