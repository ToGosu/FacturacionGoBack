# ADR-0011 — Gateway propio en Go

**Estado:** Aceptado (2026-10-06)

**Contexto.** El Gateway puede ser un producto existente (Traefik, Kong) o código propio.

**Decisión.** Construir el Gateway en Go sobre `net/http` + `httputil.ReverseProxy` con `chi`: enrutamiento por prefijo, validación de JWT, autorización por rol, límite de peticiones y `X-Correlation-Id`. Un proxy inverso aparte (Caddy o Traefik) solo termina TLS en el despliegue público.

**Consecuencias.**
- (+) Es un ejercicio valioso de Go (middleware, proxy, concurrencia) y se explica bien en entrevistas.
- (−) Más código propio que mantener que con un producto listo.

**Alternativas.** Traefik o Kong como Gateway (menos código, menos aprendizaje).
