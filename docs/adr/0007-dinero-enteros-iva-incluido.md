# ADR-0007 — Dinero y cantidades como enteros; IVA incluido en el precio

**Estado:** Aceptado (2026-10-06). Las reglas tributarias se validan con un contador.

**Contexto.** Los precios ya incluyen IVA (D-03). Los errores de redondeo con `float` son inaceptables en facturación.

**Decisión.** Dinero en **enteros de centavos** (`int64`); cantidades en **milésimas** (`int64`); tarifas en **puntos básicos**. El IVA se obtiene del total por línea (`base = round(total × 10 000 / (10 000 + tarifa))`, `iva = total − base`), con redondeo half-up y el descuento aplicado por línea. Las reglas exactas están en [`03-modelo-datos.md`](../arquitectura/03-modelo-datos.md), sección 6.

**Consecuencias.**
- (+) El total que ve el cliente nunca cambia por redondeos.
- (+) Compatible con una futura facturación electrónica.
- (−) Hay que convertir en los bordes (JSON y presentación).

**Alternativas.** `decimal` (`shopspring/decimal`) · `float` (descartado).
