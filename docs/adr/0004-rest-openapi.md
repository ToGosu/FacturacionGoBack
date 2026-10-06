# ADR-0004 — REST/JSON con contrato OpenAPI para la comunicación síncrona

**Estado:** Aceptado (2026-10-06)

**Contexto.** El tráfico síncrono es el de usuario a través del Gateway. Entre servicios casi todo es por eventos.

**Decisión.** REST/JSON con **OpenAPI escrito antes del código** (contract-first) para cada servicio, usando `chi` como router. gRPC queda fuera de v1.0 y se reevalúa como mejora opcional.

**Consecuencias.**
- (+) Simple de probar (Swagger, Postman) y de consumir desde el frontend futuro.
- (−) Se renuncia, por ahora, a la tipificación fuerte de protobuf.

**Alternativas.** gRPC entre servicios (más peso técnico, pero hay muy pocas llamadas internas que lo justifiquen) · `gin` en lugar de `chi` (equivalente para este alcance; `chi` es más cercano a la librería estándar).
