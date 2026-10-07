package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

// Servicio expone los casos de uso de escritura del catálogo (RF-CAT-01 a 04).
type Servicio struct {
	uow     UnidadDeTrabajo
	ahora   func() time.Time
	nuevoID func() string
}

// NuevoServicio crea el servicio. "ahora" y "nuevoID" son funciones y no
// interfaces: cuando una dependencia tiene un solo método, en Go lo más simple es
// recibir una función. En main.go se pasará time.Now y un generador de UUID v7;
// en las pruebas, funciones que devuelven valores fijos.
func NuevoServicio(uow UnidadDeTrabajo, ahora func() time.Time, nuevoID func() string) *Servicio {
	return &Servicio{uow: uow, ahora: ahora, nuevoID: nuevoID}
}

// CrearProductoVendible es el comando para crear un producto que se factura.
type CrearProductoVendible struct {
	Nombre      string
	TarifaIVA   domain.TarifaIVA
	Precio      domain.Centavos
	CategoriaID *domain.CategoriaID
	UsuarioID   UsuarioID
}

// CrearProductoVendible crea el ítem y registra el evento catalogo.item.actualizado.
func (s *Servicio) CrearProductoVendible(ctx context.Context, cmd CrearProductoVendible) (*domain.Item, error) {
	return s.crear(ctx, cmd.CategoriaID, cmd.UsuarioID, func(id domain.ItemID, ahora time.Time) (*domain.Item, error) {
		return domain.NuevoProductoVendible(id, cmd.Nombre, cmd.TarifaIVA, cmd.Precio, cmd.CategoriaID, ahora)
	})
}

// CrearMateriaPrima es el comando para crear una materia prima (sin precio, D-15).
type CrearMateriaPrima struct {
	Nombre      string
	Unidad      domain.Unidad
	CategoriaID *domain.CategoriaID
	UsuarioID   UsuarioID
}

// CrearMateriaPrima crea el ítem y registra el evento catalogo.item.actualizado.
func (s *Servicio) CrearMateriaPrima(ctx context.Context, cmd CrearMateriaPrima) (*domain.Item, error) {
	return s.crear(ctx, cmd.CategoriaID, cmd.UsuarioID, func(id domain.ItemID, ahora time.Time) (*domain.Item, error) {
		return domain.NuevaMateriaPrima(id, cmd.Nombre, cmd.Unidad, cmd.CategoriaID, ahora)
	})
}

// EditarItem es el comando de edición parcial (PATCH): un campo nil no cambia.
type EditarItem struct {
	ItemID    domain.ItemID
	Nombre    *string
	TarifaIVA *domain.TarifaIVA
	// CambiarCategoria indica si se toca la categoría; con CategoriaID nil se quita.
	CambiarCategoria bool
	CategoriaID      *domain.CategoriaID
	UsuarioID        UsuarioID
}

// EditarItem aplica los cambios indicados; si uno falla, no se guarda ninguno.
func (s *Servicio) EditarItem(ctx context.Context, cmd EditarItem) (*domain.Item, error) {
	return s.modificar(ctx, cmd.ItemID, cmd.UsuarioID, func(ctx context.Context, r Repositorios, it *domain.Item) error {
		if cmd.Nombre != nil {
			if err := it.Renombrar(*cmd.Nombre); err != nil {
				return err
			}
		}
		if cmd.TarifaIVA != nil {
			if err := it.CambiarTarifaIVA(*cmd.TarifaIVA); err != nil {
				return err
			}
		}
		if cmd.CambiarCategoria {
			if err := verificarCategoria(ctx, r.Categorias, cmd.CategoriaID); err != nil {
				return err
			}
			it.AsignarCategoria(cmd.CategoriaID)
		}
		return nil
	})
}

// CambiarPrecio es el comando para registrar un precio nuevo (RF-CAT-02).
type CambiarPrecio struct {
	ItemID    domain.ItemID
	Precio    domain.Centavos
	UsuarioID UsuarioID
}

// CambiarPrecio cierra el precio vigente, abre el nuevo y lo audita (RN-20).
func (s *Servicio) CambiarPrecio(ctx context.Context, cmd CambiarPrecio) (*domain.Item, error) {
	return s.modificar(ctx, cmd.ItemID, cmd.UsuarioID, func(ctx context.Context, r Repositorios, it *domain.Item) error {
		cerrado, err := it.CambiarPrecio(cmd.Precio, s.ahora())
		if err != nil {
			return err
		}
		nuevo, _ := it.PrecioVigente() // siempre existe tras un cambio exitoso
		if err := r.Items.GuardarCambioPrecio(ctx, it.ID(), cerrado, nuevo); err != nil {
			return fmt.Errorf("guardar historial de precios: %w", err)
		}
		err = r.Auditoria.Registrar(ctx, RegistroAuditoria{
			UsuarioID: cmd.UsuarioID,
			Accion:    AccionCambiarPrecio,
			Entidad:   EntidadItem,
			EntidadID: string(it.ID()),
			Antes:     map[string]any{"precio_centavos": int64(cerrado.Valor)},
			Despues:   map[string]any{"precio_centavos": int64(nuevo.Valor)},
		})
		if err != nil {
			return fmt.Errorf("registrar auditoría: %w", err)
		}
		return nil
	})
}

// DesactivarItem desactiva un ítem sin borrarlo (RF-CAT-03).
func (s *Servicio) DesactivarItem(ctx context.Context, id domain.ItemID, actor UsuarioID) (*domain.Item, error) {
	return s.modificar(ctx, id, actor, func(_ context.Context, _ Repositorios, it *domain.Item) error {
		return it.Desactivar()
	})
}

// CrearCategoria es el comando para crear una categoría (RF-CAT-04).
type CrearCategoria struct {
	Nombre string
}

// CrearCategoria crea la categoría; el nombre debe ser único.
func (s *Servicio) CrearCategoria(ctx context.Context, cmd CrearCategoria) (*domain.Categoria, error) {
	var creada *domain.Categoria
	err := s.uow.Ejecutar(ctx, func(ctx context.Context, r Repositorios) error {
		c, err := domain.NuevaCategoria(domain.CategoriaID(s.nuevoID()), cmd.Nombre)
		if err != nil {
			return err
		}
		if err := r.Categorias.Crear(ctx, c); err != nil {
			return fmt.Errorf("guardar categoría: %w", err)
		}
		creada = c
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("crear categoría: %w", err)
	}
	return creada, nil
}

// ObtenerItem devuelve un ítem por su ID o ErrItemNoEncontrado.
func (s *Servicio) ObtenerItem(ctx context.Context, id domain.ItemID) (*domain.Item, error) {
	var it *domain.Item
	err := s.uow.Ejecutar(ctx, func(ctx context.Context, r Repositorios) error {
		var err error
		it, err = r.Items.BuscarPorID(ctx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("obtener ítem %s: %w", id, err)
	}
	return it, nil
}

// crear es el flujo común de alta: valida la categoría, construye el ítem con un
// ID y hora nuevos, lo guarda y registra el evento, todo en una transacción.
func (s *Servicio) crear(ctx context.Context, categoriaID *domain.CategoriaID, actor UsuarioID,
	construir func(domain.ItemID, time.Time) (*domain.Item, error),
) (*domain.Item, error) {
	var creado *domain.Item
	err := s.uow.Ejecutar(ctx, func(ctx context.Context, r Repositorios) error {
		if err := verificarCategoria(ctx, r.Categorias, categoriaID); err != nil {
			return err
		}
		it, err := construir(domain.ItemID(s.nuevoID()), s.ahora())
		if err != nil {
			return err
		}
		if err := r.Items.Crear(ctx, it); err != nil {
			return fmt.Errorf("guardar ítem: %w", err)
		}
		if err := r.Eventos.ItemActualizado(ctx, it, actor); err != nil {
			return fmt.Errorf("registrar evento: %w", err)
		}
		creado = it
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("crear ítem: %w", err)
	}
	return creado, nil
}

// modificar es el flujo común de cambio: carga el ítem, aplica "cambiar", lo
// guarda (sube su versión) y registra el evento, todo en una transacción.
func (s *Servicio) modificar(ctx context.Context, id domain.ItemID, actor UsuarioID,
	cambiar func(context.Context, Repositorios, *domain.Item) error,
) (*domain.Item, error) {
	var modificado *domain.Item
	err := s.uow.Ejecutar(ctx, func(ctx context.Context, r Repositorios) error {
		it, err := r.Items.BuscarPorID(ctx, id)
		if err != nil {
			return err
		}
		if err := cambiar(ctx, r, it); err != nil {
			return err
		}
		if err := r.Items.Actualizar(ctx, it); err != nil {
			return fmt.Errorf("guardar ítem: %w", err)
		}
		if err := r.Eventos.ItemActualizado(ctx, it, actor); err != nil {
			return fmt.Errorf("registrar evento: %w", err)
		}
		modificado = it
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("modificar ítem %s: %w", id, err)
	}
	return modificado, nil
}

// verificarCategoria comprueba que la categoría exista; nil significa "sin categoría".
func verificarCategoria(ctx context.Context, repo CategoriaRepository, id *domain.CategoriaID) error {
	if id == nil {
		return nil
	}
	existe, err := repo.Existe(ctx, *id)
	if err != nil {
		return fmt.Errorf("verificar categoría %s: %w", *id, err)
	}
	if !existe {
		return fmt.Errorf("%w: %s", ErrCategoriaNoEncontrada, *id)
	}
	return nil
}
