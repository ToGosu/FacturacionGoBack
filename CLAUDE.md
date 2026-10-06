# CLAUDE.md — Sistema de Facturación e Inventario (Microservicios en Go)

Guía para Claude Code al trabajar en este repositorio. Léela completa antes de cada tarea.

## 1. Contexto del proyecto

- **Dominio:** facturación e inventario para una panadería (catálogo, inventario, clientes, facturación, encargos, notificaciones, reportes).
- **Propósito:** proyecto de portafolio para demostrar dominio de Go en backend. La calidad del código, la documentación y el historial de Git **son parte del entregable**.
- **Arquitectura:** microservicios donde se justifique, cada uno con **arquitectura hexagonal** (puertos y adaptadores) por dentro. No todo tiene que ser un microservicio: si una funcionalidad no justifica un servicio propio, se propone como módulo dentro de uno existente y se registra la decisión en un ADR.
- **Orden de trabajo:** primero backend (lógica de negocio); el frontend viene después.
- **Autor:** Santiago Torres está aprendiendo Go y viene de Java/Spring Boot. Cuando introduzcas un concepto propio de Go (interfaces implícitas, `error` como valor, `context.Context`, goroutines/channels, embedding), explícalo brevemente y compáralo con su equivalente en Java si ayuda. No escribas código "mágico" sin explicarlo.

## 2. Flujo de Git — OBLIGATORIO

### Regla principal: una rama por módulo o funcionalidad

**Antes de escribir cualquier código para un módulo o funcionalidad nueva, crea una rama dedicada y publícala en el repositorio remoto.** Nunca trabajes directamente sobre `main` ni `develop`.

```bash
git checkout develop
git pull origin develop
git checkout -b feature/<servicio>-<funcionalidad>
git push -u origin feature/<servicio>-<funcionalidad>
```

Si la rama ya existe en el remoto, no la recrees: haz `git fetch` y `git checkout` de esa rama.

### Ramas permanentes

| Rama | Uso |
|---|---|
| `main` | Código estable y desplegable. Solo recibe merges desde `develop` (o `hotfix/*`). Cada merge se etiqueta con versión (`v0.1.0`, …). |
| `develop` | Rama de integración. Todas las ramas de funcionalidad salen de aquí y vuelven aquí. |

### Convención de nombres de ramas

Formato: `<tipo>/<servicio>-<descripcion-corta>` en minúsculas y con guiones.

| Tipo | Cuándo | Ejemplo |
|---|---|---|
| `feature/` | Módulo o funcionalidad nueva | `feature/productos-inventario-catalogo-crud`, `feature/productos-inventario-reserva-stock` |
| `fix/` | Corrección de un bug | `fix/ventas-calculo-iva` |
| `refactor/` | Cambio interno sin alterar comportamiento | `refactor/productos-inventario-repositorio-sqlc` |
| `docs/` | Documentación, ADRs, diagramas | `docs/adr-0012-mensajeria` |
| `chore/` | Infraestructura, CI, Docker, dependencias | `chore/infra-ci-golangci-lint` |
| `test/` | Solo pruebas | `test/ventas-casos-anulacion` |
| `hotfix/` | Arreglo urgente sobre `main` | `hotfix/gateway-cors` |

Servicios válidos para el nombre (ADR-0001): `gateway` (incluye identidad), `productos-inventario`, `ventas`, `notificaciones`, `reportes`, `shared` (código común en `pkg/`), `infra`. Cuando ayude, se agrega el módulo interno después del servicio: `feature/ventas-pedidos-anticipos`.

### Granularidad

- Una rama = una funcionalidad coherente y revisable (idealmente < 400 líneas de cambio). Si un servicio es grande, divídelo: `feature/ventas-facturacion-dominio`, `feature/ventas-facturacion-api-rest`, `feature/ventas-facturacion-publicar-eventos`.
- No mezcles en una rama cambios de varios servicios, salvo que la funcionalidad lo exija (p. ej. un contrato de evento compartido); en ese caso, explícalo en el PR.

### Commits

Usa **Conventional Commits**, en español, en modo imperativo:

```
feat(inventario): agregar reserva temporal de stock
fix(facturacion): corregir redondeo del IVA en líneas con descuento
test(catalogo): cubrir validación de precio negativo
docs(adr): registrar decisión de usar NATS
```

- Commits pequeños y atómicos; cada commit debe compilar (`go build ./...`).
- Haz `git push` al remoto al terminar cada bloque de trabajo, no solo al final.
- Nunca hagas `git push --force` sobre `main` o `develop`. En ramas propias, solo `--force-with-lease` y avisando.

### Cierre de una funcionalidad (Pull Request)

1. Verifica localmente: `gofmt`, `go vet ./...`, `golangci-lint run`, `go test ./... -race`.
2. Actualiza la rama con `develop` (`git pull --rebase origin develop`) y resuelve conflictos.
3. Abre un PR hacia `develop` con `gh pr create`. La descripción debe incluir: qué se hizo, por qué, cómo probarlo y checklist de la Definición de Terminado (sección 7).
4. **No hagas merge sin la aprobación de Santiago.** Tu trabajo termina en el PR abierto.
5. Tras el merge, la rama se puede borrar del remoto.

## 3. Estructura del repositorio (monorepo)

```
.
├── CLAUDE.md
├── README.md
├── docs/
│   ├── requerimientos/        # requerimientos v1.0 (fuente de verdad) y cuestionario
│   ├── arquitectura/          # arquitectura (C4, flujos) y modelo de datos
│   ├── adr/                   # 0001-titulo.md, 0002-...
│   └── api/                   # contratos de eventos y OpenAPI (.yaml) por servicio
├── services/                  # los 5 servicios de ADR-0001
│   ├── gateway/               # módulos: identidad, gateway
│   ├── productos-inventario/  # módulos: catalogo, inventario
│   ├── ventas/                # módulos: clientes, facturacion, pedidos
│   ├── notificaciones/
│   └── reportes/
├── pkg/                       # utilidades SIN dominio: logging, config, httpx, natsx, outbox
├── deploy/                    # docker-compose, init de Postgres, NATS, observabilidad
├── .github/workflows/         # CI
├── go.work                    # une los módulos para desarrollo local
└── Makefile
```

Ruta de módulo Go: `github.com/ToGosu/FacturacionGoBack/services/<servicio>` (y `.../pkg` para el código común).

### Estructura interna de cada servicio (hexagonal)

Cada servicio agrupa uno o más **módulos** (bounded contexts). Cada módulo tiene su propio hexágono:

```
services/<servicio>/
├── cmd/<servicio>/main.go     # composición: conecta adaptadores con puertos; graceful shutdown
├── internal/
│   ├── <modulo>/              # p. ej. catalogo, inventario
│   │   ├── domain/            # entidades, value objects, reglas de negocio, errores de dominio
│   │   ├── app/               # casos de uso; define los PUERTOS (interfaces)
│   │   └── adapters/
│   │       ├── http/          # handlers REST con chi (adaptador de entrada)
│   │       ├── nats/          # publicadores/consumidores de eventos
│   │       └── postgres/      # repositorios con sqlc + pgx (adaptador de salida)
│   └── platform/              # transversal del servicio: config, auth, outbox, auditoría
├── migrations/                # migraciones SQL versionadas (golang-migrate)
├── api/openapi.yaml           # contrato del servicio (contract-first)
├── Dockerfile
└── go.mod                     # un módulo Go por servicio
```

Los módulos de un mismo servicio se hablan a través de interfaces, nunca tocando las tablas del otro, para poder separarlos después (ADR-0001). `catalogo`, `clientes` y `notificaciones` pueden ser más planos (casi CRUD) sin romper la regla de dependencias (ADR-0002).

**Reglas de dependencia (no negociables):**

- `domain` no importa nada del proyecto ni librerías de infraestructura (ni HTTP, ni SQL, ni JSON tags si se puede evitar).
- `app` depende solo de `domain` y declara las interfaces (puertos) que necesita.
- `adapters` implementan esos puertos y dependen de `app`/`domain`, nunca al revés.
- Las interfaces se definen **donde se consumen**, no donde se implementan (idioma de Go).
- Ningún servicio accede a la base de datos de otro (database-per-service). La comunicación es por API o eventos.

## 4. Stack y convenciones técnicas

- **Go:** versión estable más reciente declarada en `go.mod`. Formato con `gofmt`/`goimports`.
- **HTTP:** `chi` (preferido por ser cercano a `net/http`). No mezclar routers entre servicios.
- **Base de datos:** PostgreSQL, una instancia con una base y un usuario por servicio. Migraciones con `golang-migrate`. Acceso con `sqlc` + `pgx` (ADR-0006).
- **Mensajería:** NATS JetStream (ADR-0003), detrás de puertos para no acoplar el dominio al broker. Outbox transaccional y consumidores idempotentes (ADR-0005).
- **Autenticación:** JWT con firma asimétrica verificado por cada servicio (ADR-0008).
- **Logs:** `log/slog` (estructurado, JSON). Nunca `fmt.Println` en código de producción.
- **Configuración:** variables de entorno; nada de credenciales en el código. Mantén un `.env.example` actualizado; `.env` va en `.gitignore`.
- **Testing:** `testing` + `testify`. Tests de tabla para reglas de dominio. Tests de integración con Postgres real vía `testcontainers-go` o docker compose.

### Idiomas de Go a respetar

- Manejo de errores explícito; envolver con contexto: `fmt.Errorf("reservar stock del producto %s: %w", id, err)`. Comparar con `errors.Is/As`. Nunca ignorar un `error` con `_` sin justificarlo en un comentario.
- `context.Context` como primer parámetro en toda operación de I/O; respetar cancelación y timeouts.
- Sin `panic` para flujo normal; solo para errores irrecuperables en arranque.
- Toda goroutine debe tener un mecanismo claro de finalización (context, `WaitGroup`, `errgroup`). Apagado ordenado (graceful shutdown) en cada `main.go`.
- Nombres en Go idiomático (`ProductoID`, no `productoId`); identificadores de dominio en español, términos técnicos de Go en inglés está bien.
- Dinero: nunca `float64`. Enteros en centavos (`int64`), cantidades en milésimas y tarifas en puntos básicos; el IVA viene incluido en el precio (ADR-0007).

## 5. Reglas de negocio y contratos

- Los requerimientos aprobados (`docs/requerimientos/requerimientos-v1.0.md`) son la fuente de verdad; los ADRs de `docs/adr/` prevalecen sobre este archivo si hubiera diferencias. Si una tarea contradice o no está cubierta por ellos, **detente y pregunta**; no inventes reglas de negocio.
- Toda API nueva o modificada se documenta primero en `docs/api/` (contract-first) y luego se implementa.
- Los eventos (p. ej. `FacturaEmitida`, `StockBajo`) tienen esquema versionado y documentado. Los consumidores deben ser **idempotentes**.
- Cambios de esquema de BD solo mediante una migración nueva; nunca editar una migración ya aplicada en `develop`.

## 6. Decisiones de arquitectura (ADRs)

Toda decisión con alternativas reales (broker de mensajería, ORM vs SQL tipado, REST vs gRPC, separar o no un servicio) se registra en `docs/adr/NNNN-titulo.md` con: Contexto, Decisión, Alternativas consideradas, Consecuencias. Los ADRs van en su propia rama `docs/adr-NNNN-...` o dentro de la rama de la funcionalidad que los motiva.

## 7. Definición de Terminado (por funcionalidad)

- [ ] Rama creada desde `develop` y publicada en el remoto con el nombre correcto
- [ ] Código compila y pasa `go vet` y `golangci-lint`
- [ ] Pruebas unitarias del dominio y casos de uso; `go test ./... -race` en verde
- [ ] Contrato de API/evento documentado y consistente con la implementación
- [ ] Migraciones incluidas si aplica
- [ ] Logs estructurados y errores con contexto
- [ ] README del servicio actualizado (cómo correrlo, variables de entorno, endpoints)
- [ ] Dockerfile funcional y servicio integrado en `deploy/docker-compose.yml`
- [ ] PR abierto hacia `develop` con descripción completa

## 8. Comandos útiles

```bash
make run SERVICE=productos-inventario  # levantar un servicio
make test                        # tests de todos los servicios
make lint                        # golangci-lint
docker compose -f deploy/docker-compose.yml up -d
```

(Si el `Makefile` aún no existe, créalo en una rama `chore/infra-...` antes de depender de él.)

## 9. Cómo debe trabajar Claude en este repo

1. Lee este archivo, el README y los documentos de `docs/` relevantes a la tarea.
2. Identifica el módulo/funcionalidad y **crea y publica la rama** (sección 2) antes de editar.
3. Si la tarea es ambigua o toca reglas de negocio no definidas, pregunta antes de implementar.
4. Implementa de adentro hacia afuera: dominio → casos de uso → adaptadores → `main.go`.
5. Escribe pruebas junto con el código, no al final.
6. Commits pequeños, push frecuente, y cierra con un PR — sin hacer merge.
7. Al terminar, resume qué se hizo, qué conceptos de Go nuevos aparecieron y qué queda pendiente.
8. No agregues dependencias nuevas sin mencionarlo y justificarlo; prefiere la librería estándar.
9. No modifiques archivos de otros servicios fuera del alcance de la rama.
