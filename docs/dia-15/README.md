# Día 15: Conectividad Externa — Bridge Virtual `mc0` y NAT

## Objetivo

Implementar la arquitectura de red necesaria para otorgar acceso a Internet a los contenedores mediante un puente virtual (**bridge `mc0`**), el reenvío de paquetes en el kernel (`ip_forward`) y la traducción de direcciones de red (**NAT Masquerade** vía `iptables`).

Al finalizar este día debes poder explicar:

- Qué es un puente de red virtual (bridge Linux) y cómo actúa como un conmutador (switch) L2 de software.
- Por qué se requiere `sysctl net.ipv4.ip_forward=1` para que el kernel reenvíe paquetes entre interfaces.
- Cómo funciona la regla `iptables -t nat -A POSTROUTING -s 10.16.8.0/24 ! -o mc0 -j MASQUERADE`.
- El flujo de un paquete desde el contenedor (`10.16.8.2`) hacia Internet (`8.8.8.8`) y su retorno.

---

## Arquitectura de Red Completa

```text
                                       INTERNET (8.8.8.8)
                                               ^
                                               |
+-------------------------------- Host (Linux) | -------------------------------+
|                                              v                                |
|                           Interfaz Física (eth0 / wlan0)                      |
|                                       IP: 192.168.1.50                        |
|                                              ^                                |
|                                              | (iptables NAT MASQUERADE)      |
|                                              v                                |
|                                   Bridge Virtual: mc0                         |
|                                       IP: 10.16.8.1/24                        |
|                                              ^                                |
|                    +-------------------------+-------------------------+      |
|                    |                                                   |      |
|               veth-c1 (Host)                                      veth-c2     |
|                    ^                                                   ^      |
+--------------------|---------------------------------------------------|------+
                     | (veth pair)                                       |
+--------------------|--------------------+             +----------------|----+
| Contenedor 1       v                    |             | Contenedor 2   v    |
|               vethg-c1                  |             |               vethg |
|           IP: 10.16.8.2/24              |             |           10.16.8.3 |
|   Default Gateway: 10.16.8.1            |             |                     |
+-----------------------------------------+             +---------------------+
```

---

## Flujo de Implementación en `internal/network/bridge_linux.go`

1. **Creación del Bridge:**
   `ip link add name mc0 type bridge && ip addr add 10.16.8.1/24 dev mc0 && ip link set mc0 up`.
2. **Conectar extremo host al Bridge:**
   `ip link set veth-<id> master mc0`.
3. **Habilitar enrutamiento y NAT:**
   - `net.ipv4.ip_forward = 1` en `/proc/sys/net/ipv4/ip_forward`.
   - Regla iptables: `iptables -t nat -A POSTROUTING -s 10.16.8.0/24 ! -o mc0 -j MASQUERADE`.
4. **Enrutar dentro del Contenedor:**
   - Asignar ruta por defecto: `ip route add default via 10.16.8.1 dev vethg-<id>`.

---

## Uso desde la CLI (`mc run --nat`)

```bash
# Iniciar contenedor con conexión a Internet vía bridge mc0 y NAT
sudo ./mc run --nat /tmp/alpine-rootfs /bin/sh
```

Dentro del contenedor:
```sh
/ # ping -c 3 8.8.8.8
# PING 8.8.8.8 (8.8.8.8): 56 data bytes
# 64 bytes from 8.8.8.8: seq=0 ttl=115 time=14.2 ms
# 64 bytes from 8.8.8.8: seq=1 ttl=115 time=13.9 ms
# 64 bytes from 8.8.8.8: seq=2 ttl=115 time=14.1 ms
# 3 packets transmitted, 3 packets received, 0% packet loss
```

---

## Checklist de Cierre

- [x] Implementadas las funciones `SetupBridge`, `AttachToBridge`, `EnableNAT` y `SetupDefaultGateway`.
- [x] Creados stubs multiplataforma en `bridge_others.go` y tests unitarios en `bridge_test.go`.
- [x] Agregada la bandera `--nat` en `cmd/mc/main.go` y su propagación en `internal/namespaces`.
- [x] Documentada la topología de red completa con diagrama y flujo de paquetes.
