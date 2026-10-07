package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/app"
	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

const admin app.UsuarioID = "admin-1"

var (
	ahora     = time.Date(2026, 10, 7, 13, 0, 0, 0, time.UTC)
	pasteles  = domain.CategoriaID("cat-pasteles")
	ctxPrueba = context.Background()
)

// preparar crea el servicio con reloj fijo, IDs predecibles y una categoría existente.
func preparar(t *testing.T) (*app.Servicio, *memoria) {
	t.Helper()
	mem := nuevaMemoria()
	cat, err := domain.NuevaCategoria(pasteles, "Pasteles")
	require.NoError(t, err)
	mem.categorias[pasteles] = *cat

	ids := []string{"id-1", "id-2", "id-3"}
	nuevoID := func() string { // closure: captura y consume el slice ids
		id := ids[0]
		ids = ids[1:]
		return id
	}
	reloj := func() time.Time { return ahora }
	return app.NuevoServicio(mem, reloj, nuevoID), mem
}

func crearPastel(t *testing.T, s *app.Servicio) *domain.Item {
	t.Helper()
	it, err := s.CrearProductoVendible(ctxPrueba, app.CrearProductoVendible{
		Nombre: "Pastel de guayaba", TarifaIVA: domain.TarifaIVA19, Precio: 1_200_000,
		CategoriaID: &pasteles, UsuarioID: admin,
	})
	require.NoError(t, err)
	return it
}

func TestCrearProductoVendible(t *testing.T) {
	s, mem := preparar(t)

	it := crearPastel(t, s)

	assert.Equal(t, domain.ItemID("id-1"), it.ID(), "el ID lo genera el servicio")
	assert.True(t, it.CreadoEn().Equal(ahora), "la hora la da el reloj inyectado")
	assert.Contains(t, mem.items, it.ID())
	assert.Equal(t, []domain.ItemID{"id-1"}, mem.eventos, "se registra catalogo.item.actualizado")
}

func TestCrearItemErrores(t *testing.T) {
	inexistente := domain.CategoriaID("no-existe")
	casos := []struct {
		nombre      string
		cmd         app.CrearProductoVendible
		errEsperado error
	}{
		{"categoría inexistente", app.CrearProductoVendible{Nombre: "Pan", Precio: 50_000, CategoriaID: &inexistente}, app.ErrCategoriaNoEncontrada},
		{"regla de dominio", app.CrearProductoVendible{Nombre: "Pan", Precio: 0}, domain.ErrPrecioInvalido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s, mem := preparar(t)

			_, err := s.CrearProductoVendible(ctxPrueba, c.cmd)

			// errors.Is encuentra el error aunque venga envuelto con contexto (%w).
			require.ErrorIs(t, err, c.errEsperado)
			assert.Empty(t, mem.items)
			assert.Empty(t, mem.eventos)
		})
	}
}

func TestCrearMateriaPrima(t *testing.T) {
	s, mem := preparar(t)

	it, err := s.CrearMateriaPrima(ctxPrueba, app.CrearMateriaPrima{
		Nombre: "Harina", Unidad: domain.UnidadKilogramo, UsuarioID: admin,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TipoMateriaPrima, it.Tipo())
	assert.Len(t, mem.eventos, 1)
}

func TestFalloAlRegistrarEventoNoGuardaNada(t *testing.T) {
	s, mem := preparar(t)
	mem.errEventos = errors.New("outbox no disponible")

	_, err := s.CrearProductoVendible(ctxPrueba, app.CrearProductoVendible{Nombre: "Pan", Precio: 50_000})

	require.Error(t, err)
	assert.Empty(t, mem.items, "el ítem y su evento se guardan juntos o no se guardan (ADR-0005)")
}

func TestEditarItem(t *testing.T) {
	t.Run("aplica solo los campos indicados", func(t *testing.T) {
		s, mem := preparar(t)
		it := crearPastel(t, s)
		nombre := "Pastel de guayaba grande"
		tarifa := domain.TarifaIVA5

		editado, err := s.EditarItem(ctxPrueba, app.EditarItem{
			ItemID: it.ID(), Nombre: &nombre, TarifaIVA: &tarifa, CambiarCategoria: true, UsuarioID: admin,
		})

		require.NoError(t, err)
		assert.Equal(t, nombre, editado.Nombre())
		assert.Equal(t, tarifa, editado.TarifaIVA())
		assert.Nil(t, editado.CategoriaID(), "CambiarCategoria con nil quita la categoría")
		assert.Len(t, mem.eventos, 2)
	})

	t.Run("si un cambio falla no se guarda ninguno", func(t *testing.T) {
		s, mem := preparar(t)
		it := crearPastel(t, s)
		nombre := "Otro nombre"
		tarifa := domain.TarifaIVA(800)

		_, err := s.EditarItem(ctxPrueba, app.EditarItem{ItemID: it.ID(), Nombre: &nombre, TarifaIVA: &tarifa})

		require.ErrorIs(t, err, domain.ErrTarifaIVAInvalida)
		guardado := mem.items[it.ID()]
		assert.Equal(t, "Pastel de guayaba", guardado.Nombre())
		assert.Len(t, mem.eventos, 1, "solo el evento de creación")
	})

	t.Run("ítem inexistente", func(t *testing.T) {
		s, _ := preparar(t)
		_, err := s.EditarItem(ctxPrueba, app.EditarItem{ItemID: "no-existe"})
		require.ErrorIs(t, err, app.ErrItemNoEncontrado)
	})
}

func TestCambiarPrecio(t *testing.T) {
	t.Run("guarda historial, audita y publica", func(t *testing.T) {
		s, mem := preparar(t)
		it := crearPastel(t, s)

		_, err := s.CambiarPrecio(ctxPrueba, app.CambiarPrecio{ItemID: it.ID(), Precio: 1_350_000, UsuarioID: admin})

		require.NoError(t, err)
		require.Len(t, mem.precios, 2, "precio cerrado + precio nuevo")
		assert.Equal(t, domain.Centavos(1_200_000), mem.precios[0].Valor)
		assert.NotNil(t, mem.precios[0].VigenteHasta)
		assert.Equal(t, domain.Centavos(1_350_000), mem.precios[1].Valor)
		assert.Equal(t, []app.RegistroAuditoria{{
			UsuarioID: admin, Accion: app.AccionCambiarPrecio, Entidad: app.EntidadItem, EntidadID: "id-1",
			Antes:   map[string]any{"precio_centavos": int64(1_200_000)},
			Despues: map[string]any{"precio_centavos": int64(1_350_000)},
		}}, mem.auditoria)
		assert.Len(t, mem.eventos, 2)
	})

	t.Run("precio igual no deja rastro", func(t *testing.T) {
		s, mem := preparar(t)
		it := crearPastel(t, s)

		_, err := s.CambiarPrecio(ctxPrueba, app.CambiarPrecio{ItemID: it.ID(), Precio: 1_200_000, UsuarioID: admin})

		require.ErrorIs(t, err, domain.ErrPrecioSinCambio)
		assert.Empty(t, mem.precios)
		assert.Empty(t, mem.auditoria)
	})
}

func TestDesactivarItem(t *testing.T) {
	s, mem := preparar(t)
	it := crearPastel(t, s)

	desactivado, err := s.DesactivarItem(ctxPrueba, it.ID(), admin)
	require.NoError(t, err)
	assert.False(t, desactivado.Activo())

	_, err = s.DesactivarItem(ctxPrueba, it.ID(), admin)
	require.ErrorIs(t, err, domain.ErrItemInactivo)
	assert.Len(t, mem.eventos, 2, "creación + desactivación")
}

func TestCrearCategoria(t *testing.T) {
	s, mem := preparar(t)

	c, err := s.CrearCategoria(ctxPrueba, app.CrearCategoria{Nombre: "Panes"})
	require.NoError(t, err)
	assert.Contains(t, mem.categorias, c.ID())

	_, err = s.CrearCategoria(ctxPrueba, app.CrearCategoria{Nombre: "Pasteles"})
	require.ErrorIs(t, err, app.ErrCategoriaDuplicada)
}

func TestObtenerItem(t *testing.T) {
	s, _ := preparar(t)
	it := crearPastel(t, s)

	encontrado, err := s.ObtenerItem(ctxPrueba, it.ID())
	require.NoError(t, err)
	assert.Equal(t, it.Nombre(), encontrado.Nombre())

	_, err = s.ObtenerItem(ctxPrueba, "no-existe")
	require.ErrorIs(t, err, app.ErrItemNoEncontrado)
}
