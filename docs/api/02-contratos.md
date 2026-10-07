# Contratos — APIs y eventos

| | |
|---|---|
| **Versión** | 0.1 (borrador para revisión) |
| **Base** | `01-arquitectura.md`, Requerimientos v1.0 |

> **Alcance de este documento:** catálogo de endpoints y contratos de eventos completos. La especificación **OpenAPI formal** de cada servicio se escribe al **inicio de su semana de desarrollo** (diseño primero el contrato, luego el código), partiendo de las tablas de la sección 2. Así el contrato refleja lo que realmente se aprendió al construir el servicio anterior.

---

## 1. Convenciones

### 1.1 REST
- Prefijo `/api/v1`. JSON en UTF-8. Nombres de campo en `snake_case`, en español como el dominio.
- **Identificadores:** UUID (se recomienda v7, ordenable por tiempo).
- **Fechas y horas:** ISO 8601 en UTC (`2026-10-05T15:04:05Z`). La zona America/Bogotá se aplica solo al presentar.
- **Dinero:** entero en **centavos** (`precio_centavos: 250000` = $2.500,00). Nunca decimales ni `float`.
- **Cantidades:** cadena decimal con hasta 3 decimales (`"12.500"`), guardada internamente como entero en milésimas.
- **Tarifa de IVA:** entero en puntos básicos (`1900` = 19 %).
- **Errores:** formato *problem details* (RFC 7807): `type`, `title`, `status`, `detail`, `codigo` (código estable de negocio, p. ej. `STOCK_INSUFICIENTE`) y `errores[]` para validaciones por campo.
- **Paginación:** `?limite=50&cursor=...`; la respuesta incluye `siguiente_cursor`.
- **Idempotencia:** los `POST` críticos (emitir factura, registrar pago, anticipo, entrada de inventario) exigen la cabecera `Idempotency-Key`. Repetir la misma clave devuelve la respuesta original.
- **Correlación:** cabecera `X-Correlation-Id` generada en el Gateway y propagada a eventos y logs.
- **Autenticación:** `Authorization: Bearer <JWT>`.

### 1.2 Roles
`ADMIN` y `VENDEDOR`, según la matriz de la sección 3 de los requerimientos. En las tablas, la columna **Rol** indica el mínimo requerido (`A` = solo admin, `V` = admin o vendedor).

---

## 2. Endpoints por servicio

### 2.1 Gateway e Identidad
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| POST | `/auth/login` | pública | Credenciales → token de acceso + refresh |
| POST | `/auth/refresh` | pública | Renueva el token de acceso |
| POST | `/auth/logout` | V | Revoca el refresh token |
| GET | `/auth/.well-known/jwks.json` | pública | Claves públicas para verificar JWT |
| GET / POST | `/usuarios` | A | Listar / crear usuarios |
| PATCH | `/usuarios/{id}` | A | Cambiar rol, nombre o estado |
| GET | `/auditoria` | A | Auditoría de identidad |

El Gateway enruta por prefijo: `/items`, `/categorias`, `/inventario` → Productos e Inventario · `/clientes`, `/facturas`, `/encargos` → Ventas · `/notificaciones` → Notificaciones · `/reportes` → Reportes.

### 2.2 Productos e Inventario
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| GET / POST | `/items` | V (GET) / A (POST) | Buscar (`?q=`, `?tipo=`, `?activo=`) / crear ítem |
| GET / PATCH | `/items/{id}` | V (GET) / A (PATCH) | Detalle / editar |
| POST | `/items/{id}/desactivar` | A | Desactivar sin borrar |
| GET | `/items/{id}/precios` | V | Historial de precios |
| POST | `/items/{id}/precios` | A | Nuevo precio con vigencia (cierra el anterior) |
| GET / POST | `/categorias` | V / A | Listar / crear |
| GET | `/inventario/stock` | V | Físico, reservado y disponible (`?item_id=`, `?bajo_minimo=true`) |
| GET | `/inventario/movimientos` | V | Kardex (`?item_id=`, rango de fechas) |
| POST | `/inventario/entradas/produccion` | A | Entrada por producción propia |
| POST | `/inventario/entradas/compra` | A | Entrada de materia prima |
| POST | `/inventario/ajustes` | A | Ajuste manual con `motivo` obligatorio |
| PUT | `/inventario/umbrales/{item_id}` | A | Definir stock mínimo |
| GET | `/auditoria` | A | Auditoría del servicio |

### 2.3 Ventas
**Clientes**
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| GET / POST | `/clientes` | V | Buscar (`?documento=`, `?q=`) / crear |
| GET / PATCH | `/clientes/{id}` | V | Detalle / editar |

**Facturas**
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| POST | `/facturas` | V | Crear borrador |
| GET | `/facturas` | V | Listar (`?estado=`, `?desde=`, `?hasta=`, `?cliente_id=`) |
| GET | `/facturas/{id}` | V | Detalle con líneas y pagos |
| PUT | `/facturas/{id}/lineas` | V | Reemplazar líneas (solo en `BORRADOR`; valida disponibilidad) |
| PUT | `/facturas/{id}/descuento` | V | Aplicar descuento empleado/familia (auditado) |
| POST | `/facturas/{id}/emitir` | V | Asigna consecutivo y pasa a `EMITIDA`. *Idempotency-Key* |
| POST | `/facturas/{id}/pagos` | V | Registra un pago (efectivo, Nequi, Bancolombia). *Idempotency-Key* |
| POST | `/facturas/{id}/anular` | A | Anula con `motivo` |
| POST | `/facturas/{id}/notas-credito` | A | Nota crédito con líneas y `reingresa_inventario` por línea |
| GET | `/facturas/{id}/comprobante.pdf` | V | Comprobante (Should) |

**Encargos**
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| POST | `/encargos` | V | Crear encargo (cliente obligatorio) |
| GET | `/encargos` | V | Listar (`?estado=`, `?entrega_desde=`, `?entrega_hasta=`) |
| GET / PATCH | `/encargos/{id}` | V | Detalle / modificar (genera nueva solicitud de reserva) |
| POST | `/encargos/{id}/anticipos` | V | Registrar anticipo. *Idempotency-Key* |
| POST | `/encargos/{id}/listo` | V | Marcar listo para entrega |
| POST | `/encargos/{id}/entregar` | V | Genera la factura, aplica anticipos y devuelve la factura en borrador lista para cobrar el saldo |
| POST | `/encargos/{id}/cancelar` | V* | Cancela; si hay anticipos, `ADMIN` debe indicar `decision_anticipos` (devolver / parcial / retener) y `motivo` |

**Otros**
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| GET / PUT | `/configuracion` | A | Porcentaje de descuento empleado/familia, prefijo de factura |
| GET | `/auditoria` | A | Auditoría del servicio |

### 2.4 Notificaciones
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| GET | `/notificaciones` | V | Del usuario (por su rol); `?solo_no_leidas=true` |
| POST | `/notificaciones/{id}/leer` | V | Marcar como leída |
| POST | `/notificaciones/leer-todas` | V | Marcar todas |

En v1.0 el frontend consulta por *polling*; *Server-Sent Events* queda como mejora.

### 2.5 Reportes
| Método | Ruta | Rol | Descripción |
|---|---|:-:|---|
| GET | `/reportes/ventas` | A | Ventas por periodo (`?agrupar=dia\|semana\|mes&desde=&hasta=`) |
| GET | `/reportes/productos` | A | Ranking de más y menos vendidos (`?orden=&limite=`) |
| GET | `/reportes/pendientes-de-pago` | A | Facturas emitidas con saldo |
| GET | `/reportes/{nombre}/exportar` | A | `?formato=xlsx\|pdf` (Excel es la prioridad) |

Todos incluyen en la respuesta `actualizado_hasta` (momento del último evento procesado), porque los datos no son en tiempo real (RF-REP-06).

---

## 3. Eventos

### 3.1 Broker y nomenclatura
- **NATS JetStream** (ADR-003). Un *stream* por servicio productor: `CATALOGO_INVENTARIO` (`catalogo.>`, `inventario.>`) y `VENTAS` (`ventas.>`).
- **Subject:** `<dominio>.<agregado>.<evento>`, en minúsculas. Los eventos de stock añaden el id del ítem al final para permitir *deliver last per subject*.
- **Garantía:** al menos una entrega; consumidores durables con acuse (*ack*) manual; reintentos con *backoff*.
- **Versionado:** `schema_version` en el sobre. Un cambio incompatible crea un subject nuevo (`.v2`) y ambos conviven durante la migración.

### 3.2 Sobre común (todos los eventos)

```json
{
  "event_id": "0191f3a2-7c1e-7b1a-9d3e-5a4f6c2b8e10",
  "type": "ventas.factura.emitida",
  "schema_version": 1,
  "occurred_at": "2026-10-05T15:04:05Z",
  "producer": "ventas",
  "correlation_id": "0191f3a2-7c0d-7a55-8b21-1f3e9d4c7a02",
  "actor": { "usuario_id": "0191f3a1-ffee-7000-a000-123456789abc" },
  "data": { }
}
```

### 3.3 Catálogo de eventos

| Subject | Productor | Consumidores | Cuándo |
|---|---|---|---|
| `catalogo.item.actualizado` | Productos e Inventario | Ventas | Ítem creado, editado, desactivado o con precio nuevo |
| `inventario.stock.actualizado.{item_id}` | Productos e Inventario | Ventas | Cambia físico o reservado de un ítem |
| `inventario.stock.bajo` | Productos e Inventario | Notificaciones | El disponible cruza hacia abajo el mínimo |
| `inventario.stock.negativo` | Productos e Inventario | Notificaciones | El stock queda en negativo (D-01) |
| `inventario.reserva.confirmada` | Productos e Inventario | Ventas | Reserva asignada a un encargo |
| `inventario.reserva.en_espera` | Productos e Inventario | Ventas | No hay stock suficiente; encargo en lista de espera |
| `inventario.reserva.liberada` | Productos e Inventario | Ventas | Reserva liberada por cancelación o cambio |
| `ventas.factura.emitida` | Ventas | Productos e Inventario, Reportes | Factura pasa a `EMITIDA` |
| `ventas.factura.anulada` | Ventas | Productos e Inventario, Reportes | Factura anulada |
| `ventas.pago.recibido` | Ventas | Notificaciones | Se registra un pago (alerta en el sistema) |
| `ventas.factura.pagada` | Ventas | Notificaciones, Reportes | La suma de pagos completa el total (dispara el correo) |
| `ventas.nota_credito.emitida` | Ventas | Productos e Inventario, Reportes | Devolución o corrección |
| `ventas.encargo.reserva_solicitada` | Ventas | Productos e Inventario | Encargo creado o modificado |
| `ventas.encargo.cancelado` | Ventas | Productos e Inventario | Encargo cancelado |

### 3.4 Payloads (`data`)

**`catalogo.item.actualizado`**
```json
{
  "item_id": "uuid",
  "nombre": "Pan francés",
  "tipo": "PRODUCTO_VENDIBLE",
  "unidad": "und",
  "precio_centavos": 50000,
  "tarifa_iva_bp": 0,
  "activo": true,
  "version": 7
}
```

**`inventario.stock.actualizado.{item_id}`**
```json
{
  "item_id": "uuid",
  "fisico": "120.000",
  "reservado": "30.000",
  "disponible": "90.000",
  "version": 412
}
```

**`inventario.stock.bajo`** / **`inventario.stock.negativo`**
```json
{ "item_id": "uuid", "nombre": "Pan francés", "disponible": "8.000", "minimo": "10.000" }
```

**`inventario.reserva.confirmada`** / **`.en_espera`** / **`.liberada`**
```json
{
  "encargo_id": "uuid",
  "solicitud_version": 2,
  "faltantes": [ { "item_id": "uuid", "cantidad": "5.000" } ]
}
```
(`faltantes` solo aparece en `en_espera`.)

**`ventas.factura.emitida`**
```json
{
  "factura_id": "uuid",
  "numero": "PAN-000123",
  "encargo_id": null,
  "cliente": { "id": "uuid", "nombre": "María Pérez" },
  "lineas": [
    { "item_id": "uuid", "cantidad": "10.000", "precio_unitario_centavos": 50000,
      "total_linea_centavos": 500000, "base_centavos": 500000, "iva_centavos": 0 }
  ],
  "descuento_bp": 0,
  "total_centavos": 500000
}
```
`encargo_id` distinto de `null` indica que Inventario debe **consumir la reserva** y no descontar otra vez. `cliente` es opcional (venta de mostrador).

**`ventas.factura.anulada`**
```json
{ "factura_id": "uuid", "numero": "PAN-000123", "motivo": "Error en productos",
  "lineas": [ { "item_id": "uuid", "cantidad": "10.000" } ] }
```

**`ventas.pago.recibido`**
```json
{
  "pago_id": "uuid",
  "factura_id": "uuid",
  "numero": "PAN-000123",
  "metodo": "NEQUI",
  "monto_centavos": 250000,
  "saldo_centavos": 250000
}
```

**`ventas.factura.pagada`**
```json
{
  "factura_id": "uuid",
  "numero": "PAN-000123",
  "total_centavos": 500000,
  "cliente_nombre": "María Pérez",
  "cliente_correo": "maria@example.com"
}
```
`cliente_correo` solo viaja si el cliente tiene correo registrado; sin él, Notificaciones no envía nada (D-04). El evento lleva los datos necesarios para que Notificaciones **no tenga que consultar a Ventas**.

**`ventas.nota_credito.emitida`**
```json
{
  "nota_id": "uuid",
  "factura_id": "uuid",
  "total_centavos": 100000,
  "lineas": [ { "item_id": "uuid", "cantidad": "2.000", "monto_centavos": 100000,
                "reingresa_inventario": false } ]
}
```

**`ventas.encargo.reserva_solicitada`**
```json
{
  "encargo_id": "uuid",
  "solicitud_version": 2,
  "fecha_entrega": "2026-10-12T14:00:00Z",
  "lineas": [ { "item_id": "uuid", "cantidad": "24.000" } ]
}
```
Una nueva solicitud para el mismo `encargo_id` con `solicitud_version` mayor **reemplaza** a la anterior. Es lo que hace idempotente la modificación de un encargo.

**`ventas.encargo.cancelado`**
```json
{ "encargo_id": "uuid" }
```

### 3.5 Matriz productor → consumidor

| | Productos e Inventario | Ventas | Notificaciones | Reportes |
|---|:-:|:-:|:-:|:-:|
| **Productos e Inventario publica** | | ✔ | ✔ | |
| **Ventas publica** | ✔ | | ✔ | ✔ |

Notificaciones y Reportes **no publican** eventos de negocio. No hay dependencias circulares entre servicios por eventos de negocio, solo el ida y vuelta controlado entre Ventas e Inventario (solicitud de reserva ↔ respuesta).

---

## 4. Interpretación a confirmar

**C-01.** El requisito "correo de pago recibido" se interpreta así: la **alerta en el sistema** sale con cada pago (`ventas.pago.recibido`) y el **correo** sale una sola vez cuando la factura queda totalmente pagada (`ventas.factura.pagada`), para no mandar dos correos en un pago mixto.
