// Las pruebas están en el paquete domain_test (no domain): solo ven la API pública,
// igual que la verá el resto del servicio. Es prueba de caja negra.
package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ToGosu/FacturacionGoBack/services/productos-inventario/internal/catalogo/domain"
)

// Prueba de tabla: un slice de casos y un bucle con t.Run, que crea una subprueba
// con nombre por caso (equivale a @ParameterizedTest de JUnit).
func TestParseUnidad(t *testing.T) {
	casos := []struct {
		nombre      string
		entrada     string
		esperada    domain.Unidad
		fracciones  bool
		errEsperado error
	}{
		{"unidad", "und", domain.UnidadUnidad, false, nil},
		{"kilogramo", "kg", domain.UnidadKilogramo, true, nil},
		{"gramo", "g", domain.UnidadGramo, true, nil},
		{"litro", "l", domain.UnidadLitro, true, nil},
		{"mililitro", "ml", domain.UnidadMililitro, true, nil},
		{"mayúsculas no se aceptan", "KG", "", false, domain.ErrUnidadInvalida},
		{"fuera de la lista", "libra", "", false, domain.ErrUnidadInvalida},
		{"vacía", "", "", false, domain.ErrUnidadInvalida},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			u, err := domain.ParseUnidad(c.entrada)
			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.esperada, u)
			assert.Equal(t, c.fracciones, u.PermiteFracciones())
		})
	}
}

func TestNuevaTarifaIVA(t *testing.T) {
	casos := []struct {
		nombre string
		bp     int64
		valida bool
	}{
		{"0 %", 0, true},
		{"5 %", 500, true},
		{"19 %", 1900, true},
		{"19 sin puntos básicos", 19, false},
		{"8 % no vigente", 800, false},
		{"negativa", -500, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			tarifa, err := domain.NuevaTarifaIVA(c.bp)
			if !c.valida {
				require.ErrorIs(t, err, domain.ErrTarifaIVAInvalida)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.TarifaIVA(c.bp), tarifa)
		})
	}
}

func TestNuevaCategoria(t *testing.T) {
	casos := []struct {
		nombre      string
		id          domain.CategoriaID
		entrada     string
		esperado    string
		errEsperado error
	}{
		{"válida y recortada", "c1", "  Pastelería  ", "Pastelería", nil},
		{"longitud máxima en runas", "c1", strings.Repeat("ñ", domain.LongitudMaximaNombre), strings.Repeat("ñ", domain.LongitudMaximaNombre), nil},
		{"sin id", "", "Pastelería", "", domain.ErrIDVacio},
		{"nombre en blanco", "c1", "   ", "", domain.ErrNombreVacio},
		{"nombre muy largo", "c1", strings.Repeat("a", domain.LongitudMaximaNombre+1), "", domain.ErrNombreMuyLargo},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cat, err := domain.NuevaCategoria(c.id, c.entrada)
			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.id, cat.ID())
			assert.Equal(t, c.esperado, cat.Nombre())
		})
	}
}
