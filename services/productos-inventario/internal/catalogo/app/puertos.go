// Package app contiene los casos de uso del catálogo y declara los puertos
// (interfaces) que necesita. Los adaptadores (Postgres, NATS) los implementan.
//
// En Go una interfaz se cumple de forma implícita: un tipo que tenga los métodos
// la satisface, sin escribir "implements". Por eso las interfaces se declaran aquí,
// donde se usan, y no en el paquete que las implementa (al revés que en Java).
package app

import (
	"context"
	"errors"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

// Errores que los adaptadores devuelven y los casos de uso propagan.
var (
	ErrItemNoEncontrado      = errors.New("ítem no encontrado")
	ErrCategoriaNoEncontrada = errors.New("categoría no encontrada")
	ErrCategoriaDuplicada    = errors.New("ya existe una categoría con ese nombre")
)

// UsuarioID identifica a quien ejecuta la operación (viene del JWT, ADR-0008).
type UsuarioID string

// ItemRepository persiste ítems y su historial de precios.
type ItemRepository interface {
	Crear(ctx context.Context, it *domain.Item) error
	// Actualizar guarda los cambios del ítem e incrementa su versión.
	Actualizar(ctx context.Context, it *domain.Item) error
	// GuardarCambioPrecio cierra el precio anterior y agrega el nuevo al historial.
	GuardarCambioPrecio(ctx context.Context, id domain.ItemID, cerrado, nuevo domain.Precio) error
	// BuscarPorID devuelve ErrItemNoEncontrado si el ítem no existe.
	BuscarPorID(ctx context.Context, id domain.ItemID) (*domain.Item, error)
}

// CategoriaRepository persiste categorías.
type CategoriaRepository interface {
	// Crear devuelve ErrCategoriaDuplicada si ya existe una con el mismo nombre.
	Crear(ctx context.Context, c *domain.Categoria) error
	Existe(ctx context.Context, id domain.CategoriaID) (bool, error)
}

// Eventos registra eventos de integración en el outbox (ADR-0005): se guardan en
// la misma transacción que el cambio y otro proceso los publica después en NATS.
type Eventos interface {
	// ItemActualizado registra catalogo.item.actualizado (ver docs/api/02-contratos.md).
	ItemActualizado(ctx context.Context, it *domain.Item, actor UsuarioID) error
}

// Auditoria registra operaciones sensibles de forma inmutable (RN-20, ADR-0009).
type Auditoria interface {
	Registrar(ctx context.Context, r RegistroAuditoria) error
}

// Acciones auditadas por el catálogo.
const (
	AccionCambiarPrecio = "CAMBIAR_PRECIO"
	EntidadItem         = "item"
)

// RegistroAuditoria describe quién cambió qué y los valores antes y después.
type RegistroAuditoria struct {
	UsuarioID UsuarioID
	Accion    string
	Entidad   string
	EntidadID string
	Antes     map[string]any
	Despues   map[string]any
}

// Repositorios agrupa los puertos que se usan dentro de una misma transacción.
type Repositorios struct {
	Items      ItemRepository
	Categorias CategoriaRepository
	Eventos    Eventos
	Auditoria  Auditoria
}

// UnidadDeTrabajo ejecuta fn dentro de una transacción: si fn devuelve error,
// nada de lo que hizo se guarda (ni el cambio, ni el evento, ni la auditoría).
//
// Es el equivalente de @Transactional de Spring, pero explícito: el caso de uso
// pasa una función (closure) y el adaptador de Postgres decide BEGIN/COMMIT/ROLLBACK.
type UnidadDeTrabajo interface {
	Ejecutar(ctx context.Context, fn func(ctx context.Context, r Repositorios) error) error
}
