# Documento de Requerimientos — Sistema de Facturación e Inventario para Panadería

| | |
|---|---|
| **Versión** | 1.0 |
| **Fecha** | 5 de octubre de 2026 |
| **Autor** | Santiago Torres Castaño |
| **Estado** | Aprobado; decisiones de la sección 10 cerradas (queda una menor abierta: D-12) |
| **Fuente** | Cuestionario de requerimientos respondido (las referencias `[Q#.#]` apuntan a la pregunta de origen) |

---

## 1. Introducción

### 1.1 Propósito
Definir qué debe hacer el sistema, para quién y bajo qué reglas, antes de decidir cómo se construye. Este documento es la base de la Fase 2 (arquitectura) y de las historias de usuario de la Fase 3.

### 1.2 Contexto
Una panadería con **una sola sede** [Q1.3] vende productos de producción propia al mostrador y recibe **encargos para fechas futuras** [Q1.4]. Todo se paga de contado, en efectivo o por transferencia (Nequi / Bancolombia) [Q5.2, Q7.1]. El sistema es de **uso interno**: los clientes finales no tienen acceso [Q2.3]. Se modela desde cero, sin arrastrar el modelo del proyecto anterior [Q11.3].

### 1.3 Objetivo del proyecto
Sistema de microservicios en Go, documentado y desplegable, que sirva como pieza de portafolio. Se espera una carga aproximada de **200 facturas por día** [Q10.2].

> **Nota de honestidad técnica:** 200 facturas diarias las resiste sin esfuerzo un monolito en un servidor pequeño. La justificación de usar microservicios aquí es de **aprendizaje y portafolio**, no de escala. Esto debe quedar escrito en el primer ADR: reconocer el trade-off con claridad es más valioso ante un evaluador técnico que fingir que la escala lo exige.

### 1.4 Alcance por versiones

| Versión | Contenido |
|---|---|
| **v1.0** | Núcleo: identidad, catálogo, inventario, clientes, encargos, facturación con pagos mixtos, notificaciones, reportes |
| **v1.1** | Pago por QR con confirmación manual del vendedor, comprobantes en PDF, correo de pago recibido (si no entra en v1.0) |
| **Futuro** | Facturación electrónica DIAN [Q6.4], recetas y consumo automático de materia prima, multi-sede, proveedores |

### 1.5 Fuera de alcance (v1.0)
Facturación electrónica real ante la DIAN · ventas a crédito [Q5.2] · productos perecederos y vencimientos [Q4.6] · variantes de producto [Q3.2] · multi-sede · trazabilidad de proveedores [Q4.2] · recetas · pasarela de pagos real · acceso de clientes finales · vista de historial de compras por cliente [Q5.4].

---

## 2. Glosario

| Término | Definición |
|---|---|
| **Ítem** | Cualquier cosa con stock: producto vendible o materia prima |
| **Producto vendible** | Ítem de producción propia que se factura |
| **Materia prima** | Ítem comprado a proveedor; no se factura |
| **Stock físico** | Unidades realmente existentes |
| **Stock reservado** | Unidades apartadas para encargos |
| **Stock disponible** | Físico − reservado; es lo que se puede vender |
| **Encargo** | Pedido para entrega en fecha futura |
| **Reserva** | Apartado de stock ligado a un encargo |
| **En espera** | Encargo sin stock suficiente al crearse; se atiende cuando entre stock |
| **Factura emitida** | Factura con consecutivo asignado; ya afecta inventario |
| **Nota crédito** | Documento que corrige o revierte parcialmente una factura (devoluciones) |
| **Pago mixto** | Una factura pagada con varios métodos [Q7.3] |
| **Descuento empleado/familia** | Descuento porcentual especial [Q3.4] |
| **Kardex** | Historial de movimientos de un ítem |

---

## 3. Actores y permisos

Roles definidos: **Administrador** y **Vendedor** [Q2.1]. En el cuestionario no tenías claros los permisos [Q2.2], así que esta matriz se propuso y luego fue **aprobada** (D-05).

| Acción | Admin | Vendedor |
|---|:-:|:-:|
| Gestionar usuarios y roles | ✔ | ✘ |
| Crear/editar ítems y precios | ✔ | ✘ (solo lectura) |
| Consultar stock | ✔ | ✔ |
| Registrar entradas (producción / compra) | ✔ | ✘ |
| Ajustes manuales de inventario | ✔ | ✘ |
| Definir stock mínimo | ✔ | ✘ |
| Crear y editar borradores de factura | ✔ | ✔ |
| Emitir factura | ✔ | ✔ |
| Registrar pagos | ✔ | ✔ |
| Aplicar descuento empleado/familia | ✔ | ✔ (auditado) |
| Anular factura | ✔ | ✘ |
| Emitir nota crédito | ✔ | ✘ |
| Crear y gestionar encargos | ✔ | ✔ |
| Gestionar clientes | ✔ | ✔ |
| Ver notificaciones | ✔ | ✔ |
| Ver y exportar reportes | ✔ | ✘ |
| Consultar auditoría | ✔ | ✘ |

---

## 4. Módulos y límites (bounded contexts)

Cada módulo es dueño de sus datos y los demás solo acceden a ellos mediante su API o sus eventos, nunca a su base de datos. **Estos son límites lógicos:** cómo se agrupan en servicios desplegables se decide en la Fase 2 (ver la nota al final de esta sección).

| Módulo | Responsabilidad | NO hace |
|---|---|---|
| **Gateway e Identidad** | Login, usuarios, roles, emisión y validación de JWT, enrutamiento | No contiene lógica de negocio |
| **Catálogo** | Ítems, categorías, tarifa de IVA, historial de precios | No sabe cuánto stock hay |
| **Inventario** | Stock físico/reservado, movimientos, reservas, umbrales, alertas | No conoce precios ni facturas |
| **Clientes** | Datos de clientes | No guarda historial de compras |
| **Pedidos (encargos)** | Ciclo de vida del encargo, lista de espera | No factura ni cobra por sí mismo |
| **Facturación** | Facturas, descuentos, IVA, consecutivo, pagos mixtos, anulaciones, notas crédito | No modifica stock directamente: publica eventos |
| **Notificaciones** | Alertas en el sistema y correos | No decide cuándo ocurre un evento |
| **Reportes** | Vistas de lectura para ventas, ranking de productos, pendientes de pago, exportación | No modifica datos de negocio |

> **Nota de arquitectura:** son 8 módulos lógicos, pero **no se asume un servicio desplegable por módulo**. Convertir cada módulo en un servicio aumenta el costo (más contenedores, más pipelines, más contratos) y la Fase 3 pasaría de ~5-6 a ~7-8 semanas. En la Fase 2 se decidirá la agrupación, y cada servicio se estructurará con arquitectura hexagonal (puertos y adaptadores) para que un módulo pueda separarse después sin reescribirlo. **Reportes** se construye de último y, si hace falta, v1.0 se reduce a ventas por periodo + exportación a Excel.

---

## 5. Reglas de negocio

**Catálogo e inventario**
- **RN-01** Existe un único inventario (una sede).
- **RN-02** Hay dos tipos de ítem: `PRODUCTO_VENDIBLE` (entra por producción propia) y `MATERIA_PRIMA` (entra por compra). Solo los productos vendibles se facturan. [Q3.1, Q4.1]
- **RN-03** Los ítems son unidades simples con unidad de medida (und, kg, g, l…), sin variantes. [Q3.2]
- **RN-04** Los precios cambian en el tiempo y se conserva su historial. Cada línea de factura guarda el precio vigente **al momento de la venta**, de modo que un cambio posterior no altera facturas pasadas. [Q3.3]
- **RN-05** Stock disponible = stock físico − stock reservado.
- **RN-06** Cada ítem tiene un stock mínimo. Cuando el disponible baja de ese umbral se genera **una** alerta (no una por cada venta posterior). [Q4.5]

**Ventas y facturación**
- **RN-07** Venta inmediata sin stock disponible suficiente: **se bloquea**. [Q4.3]
- **RN-08** El stock se descuenta cuando la factura pasa a `EMITIDA`, no en borrador. Al anular, el stock vuelve automáticamente. [Q6.6]
- **RN-09** Estados de factura: `BORRADOR → EMITIDA → PAGADA`, y `EMITIDA/PAGADA → ANULADA`. [Q6.1] (ver D-02 y D-11)
- **RN-10** El consecutivo se asigna al emitir, es único y sin saltos (también prepara el terreno para la DIAN).
- **RN-11** Hay un solo impuesto, el IVA, con tarifa definida por ítem. Los precios de venta **ya incluyen el IVA**; en la factura se discrimina a partir del precio (base = precio / (1 + tarifa)). La regla de redondeo se define en la Fase 2. [Q6.3, D-03]
- **RN-12** El descuento empleado/familia es un porcentaje configurable aplicado a la factura. [Q3.4]
- **RN-13** Una factura puede pagarse con varios métodos (`EFECTIVO`, `NEQUI`, `BANCOLOMBIA`). Pasa a `PAGADA` cuando la suma de pagos iguala el total. En efectivo se registra el monto recibido y el cambio. [Q7.3]
- **RN-14** No hay crédito. "Cuentas por cobrar" se interpreta como **facturas emitidas con saldo pendiente**. [Q5.2, Q9.1]
- **RN-15** Una nota crédito indica, **por línea**, si el producto devuelto regresa al inventario o se descarta como merma. [Q6.5] (ver D-10)

**Encargos**
- **RN-16** Encargo con stock suficiente al crearse: se **reserva**. Sin stock suficiente: queda `EN_ESPERA`. [Q4.3, Q4.4]
- **RN-17** Cuando entra stock, los encargos en espera se intentan reservar en orden de fecha de entrega (el más próximo primero).
- **RN-18** Cancelar o modificar un encargo libera su reserva.
- **RN-19** Un encargo exige cliente identificado (nombre y contacto); la venta de mostrador no.
- **RN-23** Un encargo puede recibir uno o más anticipos, en cualquier método de pago. Al entregar se genera la factura, los anticipos se aplican como pagos y la factura queda con el saldo restante por cobrar. [D-06]
- **RN-24** Si un encargo con anticipos se cancela, el admin decide al cancelar si se devuelve total, parcialmente o se retiene, con motivo obligatorio y auditoría. [D-12]

*(Los identificadores RN son estables, no una secuencia: por eso RN-23 y RN-24 aparecen aquí.)*

**Transversales**
- **RN-20** Se audita (quién, qué, cuándo, valor anterior y nuevo): anulaciones, modificaciones de factura, ajustes de inventario, cambios de precio, descuentos, y cambios de usuarios/roles. El registro es inmutable. [Q10.3]
- **RN-21** Facturación debe seguir operando si Inventario no responde. [Q10.1] (ver D-01)
- **RN-22** Los eventos se procesan al menos una vez, por lo que los consumidores deben ser idempotentes (un evento duplicado no descuenta stock dos veces).

---

## 6. Requisitos funcionales

Prioridad: **M** = Must, **S** = Should, **C** = Could.

### 6.1 Gateway e Identidad
| ID | Requisito | P |
|---|---|:-:|
| RF-ID-01 | Como usuario quiero iniciar sesión con credenciales para acceder según mi rol | M |
| RF-ID-02 | Como admin quiero crear, desactivar y cambiar el rol de usuarios | M |
| RF-ID-03 | El gateway valida el JWT y aplica autorización por rol antes de enrutar | M |
| RF-ID-04 | Límite básico de peticiones por usuario/IP | C |

### 6.2 Catálogo
| ID | Requisito | P |
|---|---|:-:|
| RF-CAT-01 | Como admin quiero crear y editar ítems (nombre, unidad, tipo, tarifa de IVA, precio) | M |
| RF-CAT-02 | Cada cambio de precio genera un registro con vigencia; se puede consultar el historial | M |
| RF-CAT-03 | Desactivar un ítem sin borrarlo, para no romper facturas históricas | M |
| RF-CAT-04 | Organizar ítems por categorías | S |
| RF-CAT-05 | Búsqueda de ítems por nombre | M |

### 6.3 Inventario
| ID | Requisito | P |
|---|---|:-:|
| RF-INV-01 | Consultar stock físico, reservado y disponible por ítem | M |
| RF-INV-02 | Registrar entrada por producción propia | M |
| RF-INV-03 | Registrar entrada por compra de materia prima | M |
| RF-INV-04 | Ajuste manual con motivo obligatorio (solo admin, auditado) | M |
| RF-INV-05 | Kardex: historial de movimientos por ítem | M |
| RF-INV-06 | Definir stock mínimo por ítem | M |
| RF-INV-07 | Reservar y liberar stock para encargos | M |
| RF-INV-08 | Descontar stock al recibir `FacturaEmitida` y devolverlo al recibir `FacturaAnulada` | M |
| RF-INV-09 | Publicar `StockActualizado` y `StockBajo` | M |

### 6.4 Clientes
| ID | Requisito | P |
|---|---|:-:|
| RF-CLI-01 | Crear y editar clientes (nombre, documento, correo y/o celular, dirección) | M |
| RF-CLI-02 | Buscar cliente por documento o nombre | M |
| RF-CLI-03 | Validar el formato del documento | S |

### 6.5 Pedidos (encargos)
| ID | Requisito | P |
|---|---|:-:|
| RF-PED-01 | Crear encargo con cliente, ítems y fecha/hora de entrega | M |
| RF-PED-02 | Reservar stock si hay; si no, dejar el encargo `EN_ESPERA` | M |
| RF-PED-03 | Reintentar la reserva de los encargos en espera cuando entra stock | M |
| RF-PED-04 | Modificar o cancelar un encargo liberando su reserva | M |
| RF-PED-05 | Marcar un encargo como listo y entregarlo, generando la factura | M |
| RF-PED-06 | Listar encargos por fecha de entrega y estado | M |
| RF-PED-07 | Registrar uno o varios anticipos en el encargo (cualquier método de pago) | M |
| RF-PED-08 | Al entregar, aplicar los anticipos como pagos de la factura generada | M |
| RF-PED-09 | Al cancelar un encargo con anticipos, registrar su devolución o retención con motivo (solo admin) | M |

### 6.6 Facturación
| ID | Requisito | P |
|---|---|:-:|
| RF-FAC-01 | Crear y editar borradores de factura | M |
| RF-FAC-02 | Validar disponibilidad al agregar ítems y bloquear si no alcanza | M |
| RF-FAC-03 | Calcular subtotal, descuento, IVA y total | M |
| RF-FAC-04 | Emitir factura: asigna consecutivo y publica `FacturaEmitida` | M |
| RF-FAC-05 | Registrar pagos mixtos (efectivo, Nequi, Bancolombia) | M |
| RF-FAC-06 | Pasar a `PAGADA` al completar el total y publicar `PagoRecibido` | M |
| RF-FAC-07 | Aplicar descuento empleado/familia | M |
| RF-FAC-08 | Anular factura con motivo (solo admin) y publicar `FacturaAnulada` | M |
| RF-FAC-09 | Nota crédito por devolución o corrección | S |
| RF-FAC-10 | Comprobante de factura en PDF | S |
| RF-FAC-11 | Corregir una factura emitida mediante anulación + nueva factura (flujo guiado) o nota crédito; no hay edición directa | S |
| RF-FAC-12 | Generar QR de pago con el valor y confirmar el pago con un botón (v1.1) | C |

### 6.7 Notificaciones
| ID | Requisito | P |
|---|---|:-:|
| RF-NOT-01 | Alerta en el sistema de stock bajo para admin y vendedor | M |
| RF-NOT-02 | Alerta en el sistema de pago recibido para admin y vendedor | M |
| RF-NOT-03 | Correo de pago recibido al cliente, solo si tiene correo registrado; si no, no se envía nada | S |
| RF-NOT-04 | Listar notificaciones y marcarlas como leídas | M |
| RF-NOT-05 | No duplicar notificaciones ante eventos repetidos | M |

### 6.8 Reportes
| ID | Requisito | P |
|---|---|:-:|
| RF-REP-01 | Ventas por periodo (día, semana, mes) | M |
| RF-REP-02 | Productos más y menos vendidos | M |
| RF-REP-03 | Facturas emitidas con saldo pendiente | M |
| RF-REP-04 | Exportar a Excel (prioridad principal) [Q9.3] | M |
| RF-REP-05 | Exportar a PDF | S |
| RF-REP-06 | Datos actualizados periódicamente, no en tiempo real [Q9.2] | M |

---

## 7. Flujos críticos

**F1 — Venta de mostrador**
1. El vendedor crea un borrador y agrega productos; Facturación valida cada línea contra su copia local de disponibilidad.
2. Se aplica descuento si corresponde; el sistema calcula IVA y total.
3. El vendedor emite: se asigna consecutivo y se publica `FacturaEmitida`.
4. Inventario consume el evento y descuenta stock; si cruza el mínimo publica `StockBajo`.
5. Se registran uno o varios pagos; al completar el total la factura pasa a `PAGADA` y se publica `PagoRecibido`.
6. Notificaciones avisa a admin y vendedor.

**F2 — Anulación**
1. El admin anula indicando motivo; queda auditado.
2. Se publica `FacturaAnulada`; Inventario devuelve el stock.

**F3 — Encargo**
1. Se crea el encargo con cliente y fecha de entrega.
2. Pedidos solicita la reserva a Inventario: si hay stock queda reservado; si no, `EN_ESPERA`.
3. Cuando se registra una entrada de producción, Pedidos reintenta reservar los encargos en espera.
4. En cualquier momento pueden registrarse anticipos. Al entregar, se genera la factura, los anticipos se aplican como pagos y el vendedor cobra el saldo restante; el flujo continúa como F1.

---

## 8. Requisitos no funcionales

| ID | Requisito |
|---|---|
| RNF-01 | **Disponibilidad:** facturar debe seguir funcionando si Inventario está caído [Q10.1] |
| RNF-02 | **Consistencia:** eventual entre Facturación e Inventario; consumidores idempotentes |
| RNF-03 | **Rendimiento:** soportar ~200 facturas/día con holgura; objetivo propuesto de emitir una factura en menos de 500 ms (p95) en condiciones normales |
| RNF-04 | **Auditoría:** registro inmutable de operaciones sensibles (RN-20) |
| RNF-05 | **Seguridad:** JWT, contraseñas con hash seguro, secretos fuera del repositorio, HTTPS en el despliegue público |
| RNF-06 | **Datos personales:** el demo público usará **solo datos ficticios**. Los datos reales de clientes están sujetos al régimen colombiano de protección de datos (Ley 1581 de 2012) |
| RNF-07 | **Dinero:** montos en pesos colombianos almacenados como enteros o decimales exactos, nunca como `float` |
| RNF-08 | **Tiempo:** timestamps en UTC; presentación en America/Bogotá |
| RNF-09 | **Observabilidad:** logs estructurados, métricas y trazas distribuidas (Fase 4) |
| RNF-10 | **Despliegue:** local con Docker Compose y demo público [Q11.2], a evaluar en la Fase 5 (probablemente una sola VM, porque varios servicios + bases de datos + broker rara vez caben en capas gratuitas) |
| RNF-11 | **Mantenibilidad:** cada servicio con su base de datos, contratos versionados y pruebas unitarias mínimas |
| RNF-12 | **Idioma:** interfaz y mensajes en español |

---

## 9. Preparación para el futuro (sin implementarlo)

- **DIAN:** guardar documento del cliente, consecutivo sin saltos, tarifa de IVA por línea y facturas inmutables una vez emitidas. El proveedor de facturación electrónica se aislará detrás de una interfaz.
- **Pagos QR / transferencias:** definir una interfaz `ProveedorPago` en Facturación. v1.0 usa registro manual; v1.1 puede añadir QR con confirmación manual; una integración automática con un proveedor con webhooks podría enchufarse después sin tocar el resto.
- **Recetas:** el modelo de ítems ya distingue materia prima y producto vendible, que es el prerrequisito para agregar recetas más adelante.

---

## 10. Decisiones

Las decisiones del borrador 0.1 se resolvieron el 5 de octubre de 2026.

| ID | Tema | Decisión |
|---|---|---|
| D-01 | Bloquear venta sin stock vs. facturar con Inventario caído | Facturación mantiene una copia local de la disponibilidad alimentada por eventos; el descuento real es asíncrono. Si por una carrera el stock queda negativo, se permite temporalmente y se alerta al admin |
| D-02 | Modificar facturas emitidas | `BORRADOR` es editable; `EMITIDA` no se edita: se corrige con anulación + nueva factura o con nota crédito |
| D-03 | IVA | Los precios de venta incluyen IVA; la tarifa se define por ítem. Las tarifas reales de cada producto se validan con un contador |
| D-04 | Correo de pago recibido | Solo al cliente y solo si tiene correo registrado; si no, no se envía nada |
| D-05 | Permisos | Matriz de la sección 3 aprobada |
| D-06 | Encargos y anticipos | **Cambio respecto a la propuesta:** los encargos aceptan anticipos (RN-23, RF-PED-07 a 09). El tratamiento contable y tributario de un anticipo se valida con un contador |
| D-07 | Interfaz de usuario | Primero la lógica de negocio (backend, probado con Swagger y Postman); el frontend se construye después sobre ella |
| D-08 | Materia prima | Sin recetas; se registra y se ajusta manualmente |
| D-09 | Cuentas por cobrar | Facturas emitidas con saldo pendiente |
| D-10 | Devoluciones | Se decide por línea en la nota crédito: reingresa o merma |
| D-11 | Estado borrador/pendiente | Se trata como `BORRADOR`, sin efecto en stock |

**Decisión menor abierta**

| ID | Tema | Propuesta (se asume si no se responde) |
|---|---|---|
| D-12 | Cancelación de un encargo que ya tiene anticipos | El admin decide al cancelar si devuelve todo, una parte o retiene, con motivo y auditoría (RN-24) |

**Pendiente para la Fase 2:** agrupación de los 8 módulos en servicios desplegables (ADR).

---

## 11. Siguiente paso

Se pasa a la **Fase 2**: decidir la agrupación de módulos en servicios y la estructura hexagonal de cada uno, diagramas C4, contratos de eventos y APIs, modelo de datos por módulo, regla de redondeo del IVA incluido y los primeros ADRs (entre ellos, "por qué microservicios para ~200 facturas/día").
