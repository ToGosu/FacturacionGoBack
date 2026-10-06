# Cuestionario de Requerimientos — Sistema de Facturación e Inventario

Instrucciones: responde cada pregunta directamente debajo de la línea `> Respuesta:`. No hay respuestas "incorrectas" — si algo no aplica a tu caso, escribe "No aplica" o "Aún no lo sé" y lo definimos juntos. Mientras más detalle des aquí, menos ambigüedad habrá al diseñar la arquitectura.

---

## 1. Contexto general del negocio

1.1. ¿El negocio modelo es uno real que conoces (ej. la panadería que trabajaste antes) o es un negocio genérico/ficticio para el proyecto?
> Respuesta: Si- es una panaderia

1.2. ¿Qué tipo de negocio es? (retail, panadería, distribuidora, servicios, mixto, etc.)
> Respuesta: Panaderia

1.3. ¿El negocio tiene una sola sede/bodega o varias ubicaciones con inventario independiente?
> Respuesta: Una sola sede

1.4. ¿Vende solo productos físicos, o también servicios (que no afectan inventario)?
> Respuesta: Vende productos, pero tambien se hacen encargos o pedidos para el futuro

---

## 2. Roles y usuarios del sistema

2.1. ¿Qué roles existen? (ej. administrador, vendedor/cajero, encargado de bodega, cliente final con acceso propio, etc.)
> Respuesta: Administrador y vendedor

2.2. ¿Cada rol necesita permisos distintos sobre qué puede ver/hacer? Da un ejemplo de algo que el vendedor NO debería poder hacer pero el administrador sí.
> Respuesta: Aun no lo se realmente

2.3. ¿Los clientes finales tienen algún tipo de acceso al sistema (ej. ver sus facturas, hacer pedidos), o el sistema es solo de uso interno del negocio?
> Respuesta: Solo es interno del negocio

---

## 3. Catálogo / Productos

3.1. ¿Qué datos debe tener un producto? (nombre, SKU, precio, categoría, unidad de medida, impuesto aplicable, etc.)
> Respuesta: Nombre, precio, unidad, UUID, depende de que categorias se crearan, porque no se aun si se manejara materia prima o solo productos vendibles

3.2. ¿Los productos tienen variantes (tallas, colores, presentaciones) o son unidades simples?
> Respuesta: Son unidades simples

3.3. ¿Los precios pueden cambiar en el tiempo? ¿Se necesita guardar historial de precios?
> Respuesta: Si

3.4. ¿Hay descuentos o promociones (por producto, por cliente, por volumen)?
> Respuesta: Si, descuento de empleado/familia

---

## 4. Inventario

4.1. ¿Cómo se registra la entrada de inventario? (compra a proveedor, producción propia, ajuste manual, etc.)
> Respuesta: Produccion propia para productos normales, compra a proveedor como materia prima

4.2. ¿Se necesita trazabilidad de proveedores, o el inventario es "anónimo" (solo cantidades)?
> Respuesta: Solo cantidades

4.3. ¿Qué pasa si se intenta facturar un producto sin stock suficiente? ¿Se bloquea la venta, se permite con stock negativo, o se pone en espera?
> Respuesta: Si es un encargo para el futuro, se pone en espera si es de compra inmediate se bloquea

4.4. ¿Existe el concepto de "reserva" de stock (ej. mientras se procesa un pedido, ese stock queda apartado aunque aún no se facture)?
> Respuesta: Si

4.5. ¿Se necesitan alertas quando el stock baja de cierto nivel?
> Respuesta: Si

4.6. ¿Hay manejo de productos perecederos (fechas de vencimiento) o eso no aplica?
> Respuesta: No aplica

---

## 5. Clientes

5.1. ¿Qué datos se guardan de un cliente? (nombre, documento de identidad, contacto, dirección, etc.)
> Respuesta: Nombre, documento, contacto(correo y/o numero celular) y direccion

5.2. ¿Se manejan ventas a crédito, o todo es de contado?
> Respuesta: Todo es contado (efectivo y/o transferencia a nequi o cuenta bancolombia)

5.3. Si hay crédito: ¿se necesita límite de crédito, estados de cuenta, control de mora?
> Respuesta: No hay creditos

5.4. ¿Se necesita historial de compras por cliente?
> Respuesta: No

---

## 6. Facturación

6.1. ¿Qué estados puede tener una factura? (ej. borrador, emitida, pagada, anulada, vencida)
> Respuesta: pagada, emitida, anulada y borrador/pendiente

6.2. ¿Una factura puede modificarse después de emitida, o solo se puede anular y crear una nueva?
> Respuesta: Si, se puede modificar

6.3. ¿Cómo se calculan los impuestos? ¿Hay un solo impuesto (ej. IVA) o varios según el producto?
> Respuesta: hay un solo impuesto (IVA) aunque no se si se aplica ya sobre el precio del producto

6.4. ¿El sistema necesita cumplir con facturación electrónica legal (ej. DIAN en Colombia), o es un sistema interno sin ese requisito?
> Respuesta: Necesita(puede ser una futura funcionalidad ya que no es prioritaria)

6.5. ¿Se manejan notas crédito/débito (devoluciones, correcciones)?
> Respuesta: Devoluciones y/o correciones

6.6. ¿Qué pasa con el inventario cuando se anula una factura? (¿se devuelve el stock automáticamente?)
> Respuesta: se devuelve el stock de los prodcutos comprados autimaticamente

---

## 7. Pagos

7.1. ¿Qué métodos de pago se registran? (efectivo, tarjeta, transferencia, crédito interno)
> Respuesta: efectivo y transferencia nequi o bancolombia

7.2. ¿El sistema necesita integrarse con una pasarela de pago real, o los pagos son simulados/registrados manualmente?
> Respuesta: no necesita una pasarela de pago, pero trabaje en una empresa con un software que generaba un qr que se escaneaba y se mostraba de forma directa el valor de la factura y el usuario solo tenia que hacer la transaccion y el vendedor solo debia darle a un boton que cofirmaba el pago, que tan dificil es implementar ese sistema?

7.3. ¿Una factura puede pagarse en varias partes (abonos parciales)?
> Respuesta: Por diferentes metodos de pago? si (ej. pago 5000 en efectivo y otros 5000 por transferencia a nequi)

---

## 8. Notificaciones

8.1. ¿Qué eventos deberían generar una notificación? (ej. stock bajo, factura emitida, factura vencida, pago recibido)
> Respuesta: Stock bajo, pago recibido

8.2. ¿Por qué canal? (correo electrónico, solo dentro del sistema/dashboard, otro)
> Respuesta: Correo para los pagos, y dentro del sistema el stock bajo y los pagos

8.3. ¿Quién recibe cada tipo de notificación? (ej. stock bajo → encargado de bodega; factura emitida → cliente)
> Respuesta: Admin recibe stock bajo y pago recibido. El vendedor recibe stock bajo y pago recibido

---

## 9. Reportes

9.1. ¿Qué reportes son indispensables? (ej. ventas por periodo, productos más vendidos, rotación de inventario, cuentas por cobrar)
> Respuesta:Ventas por periodo, productos mas vendidos y menos vendidos, cuentas por cobrar o facturas por pagar

9.2. ¿Los reportes necesitan ser en tiempo real, o es aceptable que se generen con cierto retraso (ej. actualizados cada hora)?
> Respuesta: Es aceptable que se generen cada cierto tiempo

9.3. ¿Se necesita exportar reportes (PDF, Excel)?
> Respuesta: ambas, mas importante excel

---

## 10. Requisitos no funcionales

10.1. Si el servicio de Inventario está caído momentáneamente, ¿el sistema debe seguir permitiendo facturar (consistencia eventual) o debe bloquear la operación hasta confirmar stock (consistencia fuerte)? Esta respuesta define buena parte de la arquitectura de comunicación entre servicios.
> Respuesta: debe seguir permitiendo facturar

10.2. ¿Hay expectativas de carga? (ej. cuántas facturas por día se espera manejar — aunque sea una estimación, ayuda a dimensionar)
> Respuesta: por lo menos por dia unas 200 facturas

10.3. ¿Es importante la auditoría (quién hizo qué cambio y cuándo) en operaciones sensibles como anulación de facturas o ajustes de inventario?
> Respuesta: Si

10.4. ¿Hay algún requisito de seguridad particular más allá de autenticación/autorización estándar?
> Respuesta: No

---

## 11. Restricciones técnicas y de alcance

11.1. ¿Hay algún servicio/módulo que prefieras dejar fuera del alcance para no sobre-extender el proyecto? (ej. no se implementa un módulo de proveedores completo, solo se simula)
> Respuesta: Lo dejo a criterio tuyo claude

11.2. ¿Planeas desplegar esto en algún lugar público (ej. una nube gratuita) para mostrarlo en vivo, o el objetivo es que corra localmente con Docker Compose como demo?
> Respuesta: Si

11.3. ¿Hay algún aspecto del proyecto anterior de facturación/inventario (panadería) que quieras conservar tal cual, o este proyecto parte de cero en el modelado?
> Respuesta: Desde cero

---

*Cuando termines, pégame el documento completo (o solo las respuestas) y con eso armamos el documento de requerimientos formal y seguimos a la Fase 2.*
