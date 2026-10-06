# ADR-0010 — Monorepo con un módulo Go por servicio

**Estado:** Aceptado (2026-10-06)

**Contexto.** Se trabaja solo, con varios servicios y documentación compartida.

**Decisión.** Un repositorio con `services/*`, `pkg/`, `deploy/`, `docs/` y `go.work`; un `go.mod` por servicio. `pkg/` solo contiene utilidades **sin dominio** (logging, configuración, HTTP, mensajería, outbox).

**Consecuencias.**
- (+) Un único `clone`, un único pipeline y cambios coordinados fáciles.
- (+) Cada servicio compila y se prueba por separado.
- (−) Hay que vigilar que `pkg/` no se convierta en un dominio compartido.

**Alternativas.** Un repositorio por servicio (más fiel a equipos distintos, más fricción para una sola persona).
