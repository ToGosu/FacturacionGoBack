# ADR-0009 — Auditoría local e inmutable en cada servicio

**Estado:** Aceptado (2026-10-06)

**Contexto.** RN-20 exige saber quién hizo qué y cuándo en operaciones sensibles.

**Decisión.** Tabla `auditoria` en cada servicio, escrita **en la misma transacción** que el cambio, con el rol de la aplicación **sin** permisos de `UPDATE`/`DELETE`. Cada servicio expone `GET /auditoria` para el admin. Una vista consolidada queda como mejora futura.

**Consecuencias.**
- (+) Imposible tener un cambio sin su registro; simple de implementar.
- (−) No hay una única vista global de auditoría en v1.0.

**Alternativas.** Servicio de auditoría central alimentado por eventos (más una pieza, y la auditoría dejaría de ser transaccional).
