# Registro de Decisiones de Arquitectura (ADR)

Cada decisión con alternativas reales se registra en un archivo `NNNN-titulo.md` con el formato:
**Estado · Contexto · Decisión · Consecuencias · Alternativas consideradas**. Se mantienen cortos a propósito.

Un ADR aceptado no se edita para cambiar la decisión: se escribe uno nuevo que lo reemplace y el anterior pasa a estado *Reemplazado por ADR-NNNN*.

| ADR | Título | Estado |
|---|---|---|
| [0001](0001-cinco-servicios.md) | Cinco servicios para ~200 facturas diarias | Aceptado |
| [0002](0002-hexagonal-por-servicio.md) | Arquitectura hexagonal dentro de cada servicio | Aceptado |
| [0003](0003-nats-jetstream.md) | NATS JetStream como broker de eventos | Aceptado |
| [0004](0004-rest-openapi.md) | REST/JSON con contrato OpenAPI para la comunicación síncrona | Aceptado |
| [0005](0005-consistencia-outbox-idempotencia.md) | Consistencia: outbox, idempotencia y copia local | Aceptado |
| [0006](0006-postgres-sqlc-pgx.md) | PostgreSQL por servicio, `sqlc` + `pgx` y migraciones | Aceptado |
| [0007](0007-dinero-enteros-iva-incluido.md) | Dinero y cantidades como enteros; IVA incluido en el precio | Aceptado (validar con contador) |
| [0008](0008-jwt-asimetrico.md) | JWT con firma asimétrica verificado por cada servicio | Aceptado |
| [0009](0009-auditoria-local.md) | Auditoría local e inmutable en cada servicio | Aceptado |
| [0010](0010-monorepo-modulo-por-servicio.md) | Monorepo con un módulo Go por servicio | Aceptado |
| [0011](0011-gateway-propio.md) | Gateway propio en Go | Aceptado |
