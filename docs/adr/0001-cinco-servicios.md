# ADR-0001 — Cinco servicios para ~200 facturas diarias

**Estado:** Aceptado

**Contexto.** El sistema esperado maneja unas 200 facturas al día en una sola sede. Un monolito bien estructurado lo resolvería sin problema. El objetivo del proyecto es, además, aprender Go y demostrar diseño de sistemas distribuidos en un portafolio.

**Decisión.** Construir **5 servicios** (Gateway e Identidad, Productos e Inventario, Ventas, Notificaciones, Reportes) a partir de **8 módulos lógicos** mantenidos separados en el código. Los cortes se hacen donde la distribución aporta valor real: el flujo asíncrono Ventas → Inventario, un consumidor puro (Notificaciones), un modelo de lectura (Reportes) y la seguridad en un único punto.

**Consecuencias.**
- (+) Se demuestran eventos, resiliencia, gateway, observabilidad y separación de lectura sin el costo de 8 servicios.
- (+) Los módulos de un mismo servicio pueden separarse después sin reescribirlos.
- (−) Complejidad operativa mayor que la de un monolito (contenedores, contratos, consistencia eventual) que **no está justificada por la escala**. Se asume de forma consciente y se documenta aquí.

**Alternativas.** Monolito modular (descartado: no cumple el objetivo de portafolio) · un servicio por módulo (descartado: +2 semanas y más superficie sin más aprendizaje).
