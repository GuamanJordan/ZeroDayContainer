# Dia 1: fork, clone, exec y arbol de procesos

## Objetivo

Entender como Linux crea procesos y por que los contenedores se apoyan en `clone()` para aislar recursos del sistema.

Al cerrar este dia debes poder explicar:

- Que problema resuelve `fork()`.
- Que cambia cuando aparece `execve()`.
- Por que `clone()` es la base tecnica de namespaces.
- Como observar la secuencia real de syscalls con `strace`.
- Como lanzar un subproceso desde Go usando `os/exec`.

## Modelo mental

En Linux, un proceso nuevo normalmente aparece en dos pasos:

1. `fork()` crea un proceso hijo copiando el contexto del padre.
2. `execve()` reemplaza el programa del hijo por otro binario.

`clone()` es una variante mas flexible de `fork()`: permite decidir que partes del contexto se comparten con el hijo y que partes se separan. Esa flexibilidad es lo que permite crear namespaces.

## Diferencias clave

| Concepto | Que hace | Por que importa |
| --- | --- | --- |
| `fork()` | Crea un proceso hijo casi igual al padre. | Es la base historica del arbol de procesos Unix/Linux. |
| `execve()` | Reemplaza la imagen del proceso actual por otro programa. | Permite que el hijo deje de ser copia del padre y ejecute otro binario. |
| `clone()` | Crea un proceso o thread con control fino sobre que comparte. | Permite activar flags como `CLONE_NEWPID`, `CLONE_NEWUTS` o `CLONE_NEWNS`. |

## Demo manual con `strace`

Ejecuta:

```bash
strace -f -e trace=clone,execve,wait4,exit_group ls
```

Que debes observar:

- `execve()` carga el binario `ls`.
- Si el comando crea hijos, `strace -f` sigue esos procesos.
- El proceso termina con `exit_group()`.

Para observar un comando que si lanza un hijo de forma evidente:

```bash
strace -f -e trace=clone,execve,wait4,exit_group sh -c 'echo hijo'
```

Lectura esperada:

- `execve()` inicia `sh`.
- `sh` prepara la ejecucion del comando.
- El proceso hijo ejecuta `echo`.
- El padre espera al hijo con `wait4()`.

## Demo minima en Go

Este programa muestra como Go usa `os/exec` para lanzar un subproceso y capturar su PID:

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("sh", "-c", "echo PID del hijo: $$")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		os.Exit(1)
	}

	fmt.Println("PID observado por el padre:", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "wait:", err)
		os.Exit(1)
	}
}
```

Para ejecutarlo sin agregar codigo permanente al repo:

```bash
cat >/tmp/fork-exec-demo.go <<'GO'
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("sh", "-c", "echo PID del hijo: $$")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		os.Exit(1)
	}

	fmt.Println("PID observado por el padre:", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "wait:", err)
		os.Exit(1)
	}
}
GO

go run /tmp/fork-exec-demo.go
```

## Relacion con el proyecto

El runtime `mc` necesita lanzar procesos hijos controlados. En Go, ese punto de entrada sera `exec.Command`.

En los dias siguientes se agregan flags de `clone()` a traves de:

```go
cmd.SysProcAttr = &syscall.SysProcAttr{
	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
}
```

Ese cambio transforma un subproceso normal en un proceso con namespaces propios.

## Checklist de cierre

- [ ] Lei `man fork`.
- [ ] Lei `man execve`.
- [ ] Lei `man clone`.
- [ ] Ejecute `strace -f` sobre un comando simple.
- [ ] Ejecute la demo minima con `os/exec`.
- [ ] Puedo explicar la diferencia entre `fork`, `clone` y `exec`.

## Validacion sugerida

```bash
go build ./...
go test ./...
git diff --check
```
