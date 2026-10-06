# ADR-0005 — Consistencia: outbox, idempotencia y copia local

**Estado:** Aceptado (2026-10-06)

**Contexto.** RNF-01 exige facturar aunque Inventario esté caído, y la venta inmediata sin stock debe bloquearse (D-01).

**Decisión.**
1. **Outbox transaccional** en cada servicio que publica.
2. **Consumidores idempotentes** (`eventos_procesados` + restricciones únicas).
3. **Copia local** `item_snapshot` en Ventas (precio, IVA, activo, disponible), alimentada por eventos.
4. La **lista de espera de reservas vive en Inventario**, no en Pedidos.
5. Si por una carrera el stock queda negativo, se permite temporalmente y se alerta al admin.

**Consecuencias.**
- (+) Cumple D-01 y RNF-01 completos, también para precios.
- (+) La lista de espera evita carreras entre servicios.
- (−) Existe una ventana de segundos con datos desactualizados; es aceptable a este volumen y está documentada.
- (−) Hay que construir y probar el relay del outbox y la deduplicación (es parte del aprendizaje).

**Alternativas.** Llamada síncrona de Ventas a Inventario en cada venta (rompe RNF-01) · transacciones distribuidas / 2PC (innecesario y frágil) · saga orquestada (más piezas; la coreografiada basta con dos servicios involucrados).
