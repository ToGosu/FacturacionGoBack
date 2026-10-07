# ADR-0008 — JWT con firma asimétrica verificado por cada servicio

**Estado:** Aceptado (2026-10-06)

**Contexto.** Hay un Gateway, pero no conviene que los servicios confíen solo por estar en la red interna.

**Decisión.** Identidad emite JWT de acceso (~15 min) firmados con clave asimétrica (EdDSA o RS256) y publica las claves públicas en un endpoint JWKS. Cada servicio **verifica la firma** y lee el rol del *claim*. *Refresh tokens* guardados como hash. Contraseñas con argon2id.

**Consecuencias.**
- (+) Ningún servicio necesita llamar a Identidad por cada petición.
- (+) Defensa en profundidad: el Gateway filtra, y los servicios revalidan.
- (−) Revocar un token de acceso antes de que expire no es inmediato (aceptable con duraciones cortas).

**Alternativas.** Sesiones en servidor (acoplan los servicios a Identidad) · un proveedor externo de identidad (más peso, menos aprendizaje).
