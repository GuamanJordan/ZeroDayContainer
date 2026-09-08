# Día 14: Aislamiento de Red — Network Namespaces y Veth Pairs

## Objetivo

Comprender la virtualización de la pila de red en Linux mediante **Network Namespaces** (`CLONE_NEWNET`) y la interconexión entre el host y el contenedor utilizando un par de interfaces virtuales Ethernet (**veth pair**), permitiendo tráfico bidireccional punto a punto con direccionamiento IP privado.

Al finalizar este día debes poder explicar:

- Qué aísla un Network Namespace (tablas de enrutamiento, interfaces, reglas de firewall, sockets y puertos).
- Qué es un `veth pair` y por qué funciona conceptualmente como un cable Ethernet virtual bidireccional.
- Cómo mover un extremo del veth pair al network namespace del contenedor mientras el otro permanece en el host.
- Cómo configurar direcciones IP (`10.16.8.1/24` y `10.16.8.2/24`) y verificar conectividad con `ping`.

---

## Concepto: El Cable de Red Virtual (`veth pair`)

Un par `veth` siempre se crea de a dos: los paquetes enviados por una de las interfaces salen instantáneamente por la otra:

```text
+------------------------- Host (Default NetNS) -------------------------+
|                                                                        |
|   veth-host (IP: 10.16.8.1/24)                                         |
|        ^                                                               |
+--------|---------------------------------------------------------------+
         |  (Cable Ethernet Virtual / veth pair)
+--------|---------------------------------------------------------------+
|        v                                                               |
|   vethg-cont (IP: 10.16.8.2/24)                                        |
|                                                                        |
| +------------------- Contenedor (CLONE_NEWNET) ----------------------+ |
| | PID 1 (sh / workload)                                              | |
| | loopback 'lo' (127.0.0.1)                                          | |
| +--------------------------------------------------------------------+ |
+------------------------------------------------------------------------+
```

---

## Implementación en `internal/network`

El paquete `internal/network/veth_linux.go` realiza la siguiente secuencia:
1. **Creación del par:** `ip link add veth-<id> type veth peer name vethg-<id>`.
2. **Mover al namespace:** `ip link set vethg-<id> netns <pid_hijo>`.
3. **Configurar el host:** Asigna `10.16.8.1/24` a `veth-<id>` y la levanta con `ip link set ... up`.
4. **Configurar el contenedor:** Vía `nsenter -t <pid> -n`, asigna `10.16.8.2/24` a `vethg-<id>` y levanta `lo` y la interfaz.
5. **Limpieza automática:** Al finalizar el proceso contenedor, `defer CleanupVethPair(hostVeth)` destruye la interfaz del host, eliminando automáticamente el par entero del kernel.

---

## Uso desde la CLI (`mc run --net`)

```bash
# Ejecutar un contenedor con stack de red aislado y veth pair
sudo ./mc run --net /tmp/alpine-rootfs /bin/sh
```

Dentro del contenedor:
```sh
/ # ip addr
# 1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 ...
# 2: vethg-xxx: <BROADCAST,MULTICAST,UP,LOWER_UP> inet 10.16.8.2/24 ...

/ # ping -c 3 10.16.8.1
# PING 10.16.8.1 (10.16.8.1): 56 data bytes
# 64 bytes from 10.16.8.1: seq=0 ttl=64 time=0.082 ms
# 3 packets transmitted, 3 packets received, 0% packet loss
```

---

## Checklist de Cierre

- [x] Creado el paquete `internal/network` (`veth_linux.go`, `veth_others.go`, `veth_test.go`).
- [x] Agregada bandera `CLONE_NEWNET` mediante `GetNetworkSysProcAttr()` en `internal/namespaces`.
- [x] Implementada la asignación y levantamiento de interfaces y loopback con `nsenter`.
- [x] Integrada la bandera `--net` en `cmd/mc/main.go`.
- [x] Limpieza garantizada de interfaces virtuales al salir del contenedor.
