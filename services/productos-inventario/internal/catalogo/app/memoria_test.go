package app_test

import (
	"context"
	"maps"
	"slices"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/app"
	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

// memoria implementa todos los puertos en memoria para probar los casos de uso
// sin base de datos. Imita una transacción: Ejecutar trabaja sobre una copia del
// estado y solo la confirma si la función no devuelve error.
type memoria struct {
	estado
	errEventos error // si no es nil, registrar un evento falla (para probar rollback)
}

type estado struct {
	items      map[domain.ItemID]domain.Item // se guardan copias, no punteros
	categorias map[domain.CategoriaID]domain.Categoria
	precios    []domain.Precio
	eventos    []domain.ItemID
	auditoria  []app.RegistroAuditoria
}

func nuevaMemoria() *memoria {
	return &memoria{estado: estado{
		items:      map[domain.ItemID]domain.Item{},
		categorias: map[domain.CategoriaID]domain.Categoria{},
	}}
}

func (e estado) clonar() estado {
	return estado{
		items:      maps.Clone(e.items),
		categorias: maps.Clone(e.categorias),
		precios:    slices.Clone(e.precios),
		eventos:    slices.Clone(e.eventos),
		auditoria:  slices.Clone(e.auditoria),
	}
}

// Ejecutar cumple app.UnidadDeTrabajo.
func (m *memoria) Ejecutar(ctx context.Context, fn func(context.Context, app.Repositorios) error) error {
	tx := &transaccion{estado: m.clonar(), errEventos: m.errEventos}
	err := fn(ctx, app.Repositorios{
		Items:      itemsMem{tx},
		Categorias: categoriasMem{tx},
		Eventos:    eventosMem{tx},
		Auditoria:  auditoriaMem{tx},
	})
	if err != nil {
		return err // "rollback": la copia se descarta
	}
	m.estado = tx.estado // "commit"
	return nil
}

type transaccion struct {
	estado
	errEventos error
}

// Un tipo por puerto: Go no permite dos métodos Crear con distinta firma en el mismo tipo.
type (
	itemsMem      struct{ tx *transaccion }
	categoriasMem struct{ tx *transaccion }
	eventosMem    struct{ tx *transaccion }
	auditoriaMem  struct{ tx *transaccion }
)

func (r itemsMem) Crear(_ context.Context, it *domain.Item) error {
	r.tx.items[it.ID()] = *it
	return nil
}

func (r itemsMem) Actualizar(_ context.Context, it *domain.Item) error {
	if _, ok := r.tx.items[it.ID()]; !ok {
		return app.ErrItemNoEncontrado
	}
	r.tx.items[it.ID()] = *it
	return nil
}

func (r itemsMem) GuardarCambioPrecio(_ context.Context, _ domain.ItemID, cerrado, nuevo domain.Precio) error {
	r.tx.precios = append(r.tx.precios, cerrado, nuevo)
	return nil
}

func (r itemsMem) BuscarPorID(_ context.Context, id domain.ItemID) (*domain.Item, error) {
	it, ok := r.tx.items[id]
	if !ok {
		return nil, app.ErrItemNoEncontrado
	}
	return &it, nil // puntero a una copia: modificarla no altera lo guardado
}

func (r categoriasMem) Crear(_ context.Context, c *domain.Categoria) error {
	for _, existente := range r.tx.categorias {
		if existente.Nombre() == c.Nombre() {
			return app.ErrCategoriaDuplicada
		}
	}
	r.tx.categorias[c.ID()] = *c
	return nil
}

func (r categoriasMem) Existe(_ context.Context, id domain.CategoriaID) (bool, error) {
	_, ok := r.tx.categorias[id]
	return ok, nil
}

func (r eventosMem) ItemActualizado(_ context.Context, it *domain.Item, _ app.UsuarioID) error {
	if r.tx.errEventos != nil {
		return r.tx.errEventos
	}
	r.tx.eventos = append(r.tx.eventos, it.ID())
	return nil
}

func (r auditoriaMem) Registrar(_ context.Context, reg app.RegistroAuditoria) error {
	r.tx.auditoria = append(r.tx.auditoria, reg)
	return nil
}
