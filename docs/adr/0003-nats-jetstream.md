# ADR-0003 — NATS JetStream como broker de eventos

**Estado:** Aceptado (2026-10-06)

**Contexto.** Se necesita entrega al menos una vez, consumidores durables, reintentos y reproducción de eventos, en una máquina pequeña. Kafka tiene más peso en un CV.

**Decisión.** **NATS JetStream**. Es ligero (un binario), tiene consumidores durables, deduplicación por `Nats-Msg-Id` y *deliver last per subject* (útil para la copia local de stock). El broker queda detrás de un puerto (`PublicadorEventos` / `SuscriptorEventos`), de modo que un adaptador de Kafka es posible más adelante como ejercicio adicional.

**Consecuencias.**
- (+) Cabe en una VM pequeña junto al resto del sistema; curva de aprendizaje corta mientras se aprende Go.
- (−) Menos reconocimiento en ofertas laborales que Kafka.
- (−) Menos ecosistema de herramientas que Kafka.

**Alternativas.** Kafka (más peso en el CV, pero consume mucho más y añade operación pesada para 200 facturas/día) · RabbitMQ.

**Pregunta abierta:** si la oferta laboral menciona Kafka, vale la pena reconsiderarlo.
