# FacturacionGo — Backend

Backend en Go de un sistema de facturación e inventario para una panadería: catálogo, inventario, ventas (clientes, facturación, encargos), notificaciones y reportes.

Proyecto de portafolio: microservicios con arquitectura hexagonal, eventos con NATS JetStream y PostgreSQL por servicio. Las decisiones están en [`docs/adr/`](docs/adr/) y la guía de trabajo del repositorio en [CLAUDE.md](CLAUDE.md).

> En construcción: por ahora existe la infraestructura local; los servicios se agregan uno por uno.

## Requisitos

- Go 1.27+
- Docker (Docker Desktop en Windows)
- `make` y `golangci-lint` v2

## Puesta en marcha

```bash
cp .env.example .env   # y completa las contraseñas: compose no arranca sin ellas
make up                # Postgres, NATS JetStream y Mailpit
make ps
```

| Componente | Dirección local |
|---|---|
| PostgreSQL | `localhost:5433` (una base y un usuario por servicio) |
| NATS | `localhost:4222` · monitoreo en http://localhost:8222 |
| Mailpit (SMTP de pruebas) | SMTP `localhost:1025` · bandeja en http://localhost:8025 |

## Comandos

| Comando | Qué hace |
|---|---|
| `make build` / `make vet` | Compila / analiza todos los módulos |
| `make test` | Pruebas de todos los módulos |
| `make test-race` | Pruebas con detector de carreras, dentro de un contenedor Linux (en Windows `-race` necesita gcc) |
| `make lint` / `make fmt` | `golangci-lint` (incluye gofmt y goimports) / formatea el código |
| `make run SERVICE=<servicio>` | Ejecuta un servicio |
| `make up` / `make down` / `make logs` | Maneja la infraestructura de Docker Compose |

## Estructura

```
services/   un módulo Go por servicio (gateway, productos-inventario, ventas, notificaciones, reportes)
pkg/        utilidades sin dominio compartidas por los servicios
deploy/     docker-compose e inicialización de Postgres
docs/       requerimientos, arquitectura, ADRs y contratos
go.work     une los módulos para desarrollo local
```
