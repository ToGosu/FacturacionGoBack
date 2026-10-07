# ADR-0002 — Arquitectura hexagonal dentro de cada servicio

**Estado:** Aceptado

**Contexto.** Cada servicio mezcla reglas de negocio con HTTP, base de datos y mensajería. Hexagonal y microservicios no se excluyen: operan en niveles distintos (la primera organiza el interior de un servicio; la segunda, cuántos servicios hay).

**Decisión.** Cada módulo se organiza en `domain` / `app` / `adapters`. Las dependencias apuntan hacia adentro y los puertos los define el código que los usa. Se aplica completa en `facturacion`, `pedidos` e `inventario`; `catalogo`, `clientes` y `notificaciones` pueden ser más planos.

**Consecuencias.**
- (+) El dominio se prueba sin base de datos ni broker.
- (+) Cambiar de adaptador (otro broker, otro proveedor de pagos) no toca las reglas.
- (+) Aprovecha las interfaces implícitas de Go.
- (−) Más archivos e interfaces. Se evita abstraer de más: no se crean puertos sin una segunda implementación plausible o una necesidad de prueba.

**Alternativas.** Capas clásicas handler → servicio → repositorio (más simple, pero el servicio termina dependiendo de detalles de infraestructura).
