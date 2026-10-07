package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

var (
	// Hora fija en Bogotá (UTC-5): las pruebas verifican que el dominio la guarde en UTC.
	bogota = time.FixedZone("America/Bogota", -5*60*60)
	t0     = time.Date(2026, 10, 6, 8, 0, 0, 0, bogota)
)

func nuevoPastel(t *testing.T) *domain.Item {
	t.Helper() // si falla, el error apunta a la prueba que llamó, no a esta línea
	it, err := domain.NuevoProductoVendible("i1", "Pastel de arequipe", domain.TarifaIVA19, 1_200_000, nil, t0)
	require.NoError(t, err)
	return it
}

func TestNuevoProductoVendible(t *testing.T) {
	cat := domain.CategoriaID("pasteleria")
	casos := []struct {
		nombre      string
		id          domain.ItemID
		item        string
		tarifa      domain.TarifaIVA
		precio      domain.Centavos
		errEsperado error
	}{
		{"válido", "i1", "Pastel de guayaba", domain.TarifaIVA19, 1_200_000, nil},
		{"tarifa 0 %", "i1", "Pan", domain.TarifaIVA0, 50_000, nil},
		{"sin id", "", "Pan", domain.TarifaIVA0, 50_000, domain.ErrIDVacio},
		{"nombre vacío", "i1", "  ", domain.TarifaIVA0, 50_000, domain.ErrNombreVacio},
		{"tarifa no permitida", "i1", "Pan", domain.TarifaIVA(800), 50_000, domain.ErrTarifaIVAInvalida},
		{"precio cero", "i1", "Pan", domain.TarifaIVA0, 0, domain.ErrPrecioInvalido},
		{"precio negativo", "i1", "Pan", domain.TarifaIVA0, -100, domain.ErrPrecioInvalido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			it, err := domain.NuevoProductoVendible(c.id, c.item, c.tarifa, c.precio, &cat, t0)
			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado)
				assert.Nil(t, it)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.TipoProductoVendible, it.Tipo())
			assert.Equal(t, domain.UnidadUnidad, it.Unidad(), "los vendibles se cuentan en unidades enteras (D-13)")
			assert.Equal(t, c.tarifa, it.TarifaIVA())
			assert.Equal(t, &cat, it.CategoriaID())
			assert.True(t, it.Activo())
			assert.Equal(t, time.UTC, it.CreadoEn().Location())

			precio, ok := it.PrecioVigente()
			require.True(t, ok)
			assert.Equal(t, c.precio, precio.Valor)
			assert.True(t, precio.VigenteDesde.Equal(t0))
			assert.Nil(t, precio.VigenteHasta)
		})
	}
}

func TestNuevaMateriaPrima(t *testing.T) {
	t.Run("válida, sin precio ni IVA", func(t *testing.T) {
		it, err := domain.NuevaMateriaPrima("m1", "Harina de trigo", domain.UnidadKilogramo, nil, t0)
		require.NoError(t, err)
		assert.Equal(t, domain.TipoMateriaPrima, it.Tipo())
		assert.Equal(t, domain.UnidadKilogramo, it.Unidad())
		assert.Equal(t, domain.TarifaIVA0, it.TarifaIVA())
		_, tienePrecio := it.PrecioVigente()
		assert.False(t, tienePrecio, "la materia prima no tiene precio (D-15)")
	})

	t.Run("unidad no permitida", func(t *testing.T) {
		_, err := domain.NuevaMateriaPrima("m1", "Harina", domain.Unidad("libra"), nil, t0)
		require.ErrorIs(t, err, domain.ErrUnidadInvalida)
	})
}

func TestItemCambiarPrecio(t *testing.T) {
	t1 := t0.Add(48 * time.Hour)

	t.Run("cierra el precio anterior y abre el nuevo", func(t *testing.T) {
		it := nuevoPastel(t)

		cerrado, err := it.CambiarPrecio(1_350_000, t1)

		require.NoError(t, err)
		assert.Equal(t, domain.Centavos(1_200_000), cerrado.Valor)
		require.NotNil(t, cerrado.VigenteHasta)
		assert.True(t, cerrado.VigenteHasta.Equal(t1))

		vigente, _ := it.PrecioVigente()
		assert.Equal(t, domain.Centavos(1_350_000), vigente.Valor)
		assert.True(t, vigente.VigenteDesde.Equal(t1), "rige desde que se registra (D-16)")
		assert.Nil(t, vigente.VigenteHasta)
	})

	casos := []struct {
		nombre      string
		valor       domain.Centavos
		ahora       time.Time
		errEsperado error
	}{
		{"precio cero", 0, t1, domain.ErrPrecioInvalido},
		{"mismo precio", 1_200_000, t1, domain.ErrPrecioSinCambio},
		{"antes del vigente", 1_350_000, t0.Add(-time.Minute), domain.ErrVigenciaInvalida},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			it := nuevoPastel(t)

			_, err := it.CambiarPrecio(c.valor, c.ahora)

			require.ErrorIs(t, err, c.errEsperado)
			vigente, _ := it.PrecioVigente()
			assert.Equal(t, domain.Centavos(1_200_000), vigente.Valor, "un error no modifica el ítem")
		})
	}

	t.Run("materia prima no tiene precio", func(t *testing.T) {
		it, err := domain.NuevaMateriaPrima("m1", "Azúcar", domain.UnidadKilogramo, nil, t0)
		require.NoError(t, err)

		_, err = it.CambiarPrecio(500_000, t1)

		require.ErrorIs(t, err, domain.ErrNoAplicaAMateriaPrima)
	})
}

func TestItemPrecioVigenteDevuelveCopia(t *testing.T) {
	it := nuevoPastel(t)

	precio, _ := it.PrecioVigente()
	precio.Valor = 1

	vigente, _ := it.PrecioVigente()
	assert.Equal(t, domain.Centavos(1_200_000), vigente.Valor, "modificar la copia no altera el ítem")
}

func TestItemCambiarTarifaIVA(t *testing.T) {
	t.Run("producto vendible", func(t *testing.T) {
		it := nuevoPastel(t)
		require.NoError(t, it.CambiarTarifaIVA(domain.TarifaIVA5))
		assert.Equal(t, domain.TarifaIVA5, it.TarifaIVA())
	})

	t.Run("tarifa no permitida", func(t *testing.T) {
		it := nuevoPastel(t)
		require.ErrorIs(t, it.CambiarTarifaIVA(domain.TarifaIVA(1600)), domain.ErrTarifaIVAInvalida)
		assert.Equal(t, domain.TarifaIVA19, it.TarifaIVA())
	})

	t.Run("materia prima", func(t *testing.T) {
		it, err := domain.NuevaMateriaPrima("m1", "Leche", domain.UnidadLitro, nil, t0)
		require.NoError(t, err)
		require.ErrorIs(t, it.CambiarTarifaIVA(domain.TarifaIVA5), domain.ErrNoAplicaAMateriaPrima)
	})
}

func TestItemRenombrarYCategoria(t *testing.T) {
	it := nuevoPastel(t)

	require.NoError(t, it.Renombrar("  Pastel de arequipe grande "))
	assert.Equal(t, "Pastel de arequipe grande", it.Nombre())
	require.ErrorIs(t, it.Renombrar(""), domain.ErrNombreVacio)
	assert.Equal(t, "Pastel de arequipe grande", it.Nombre(), "un nombre inválido no se aplica")

	cat := domain.CategoriaID("tortas")
	it.AsignarCategoria(&cat)
	assert.Equal(t, &cat, it.CategoriaID())
	it.AsignarCategoria(nil)
	assert.Nil(t, it.CategoriaID())
}

func TestItemDesactivar(t *testing.T) {
	it := nuevoPastel(t)

	require.NoError(t, it.Desactivar())
	assert.False(t, it.Activo())
	require.ErrorIs(t, it.Desactivar(), domain.ErrItemInactivo)
}
