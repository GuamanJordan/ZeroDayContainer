# Día 22: Mapeo y Reenvío de Puertos con DNAT (`-p`)

## Objetivo

Implementar el reenvío de puertos desde el host hacia el contenedor mediante **Destination NAT (DNAT)** con `iptables`, permitiendo que servicios de red que se ejecutan dentro del contenedor (servidores web HTTP, APIs, microservicios) sean accesibles tanto desde la red externa como desde el propio host a través de `localhost`.

Al finalizar este día debes poder explicar:

- La diferencia entre la cadena `PREROUTING` (paquetes entrantes por interfaces de red físicas) y la cadena `OUTPUT` (paquetes originados localmente en el host hacia `127.0.0.1`).
- Cómo el kernel Linux reescribe la cabecera IP de destino usando la acción `-j DNAT --to-destination <ip>:<puerto>`.
- El manejo de protocolos `tcp` y `udp` en las reglas de cortafuegos.
- La sintaxis y ciclo de vida de la bandera `-p` / `--port <host_port>:<container_port>[/protocol]`.

---

## Arquitectura de Reenvío de Puertos (DNAT)

```text
Cliente Externo / Navegador (ej: http://192.168.1.50:8080)
                |
                v
        [ iptables PREROUTING ]
        (Reescribe dst: 192.168.1.50:8080 -> 10.16.8.2:80)
                |
                v
        [ Bridge mc0 ] ----> [ veth-pair ] ----> [ Contenedor / Nginx (10.16.8.2:80) ]
```

### Reenvío Local (Host hacia Localhost)

Cuando un proceso en el host intenta acceder a `http://localhost:8080`, el paquete no pasa por `PREROUTING` porque se genera localmente. Para garantizar compatibilidad total con herramientas de desarrollo (e.g. `curl http://localhost:8080`), se configura una regla complementaria en la cadena `OUTPUT`:

```bash
iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 8080 -j DNAT --to-destination 10.16.8.2:80
```

---

## Sintaxis y Ejemplos en `mc`

### 1. Exponer un servidor web en el puerto 8080
```bash
sudo mc run --nat -p 8080:80 /tmp/alpine-rootfs /bin/sh
```
*Si dentro del contenedor inicias un servidor con `python3 -m http.server 80` o Nginx, puedes acceder desde tu navegador en `http://localhost:8080`.*

### 2. Exponer un servicio UDP
```bash
sudo mc run --nat -p 5353:53/udp /tmp/alpine-rootfs /bin/sh
```

---

## Checklist de Cierre

- [x] Implementado el paquete `internal/network/portmap_linux.go` con `ParsePortMapping`, `SetupPortForwarding` y `CleanupPortForwarding`.
- [x] Añadidos stubs compatibles en `internal/network/portmap_others.go`.
- [x] Pruebas unitarias en `internal/network/portmap_test.go`.
- [x] Integración de la bandera `-p` / `--port` en `cmd/mc/main.go` y `internal/namespaces/run_linux.go`.
- [x] Limpieza garantizada (`defer`) de las reglas iptables al salir del contenedor.
