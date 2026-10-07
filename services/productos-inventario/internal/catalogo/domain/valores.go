package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// LongitudMaximaNombre limita el nombre de ítems y categorías.
const LongitudMaximaNombre = 100

// Centavos representa dinero en centavos de peso colombiano (ADR-0007).
// Nunca se usa float64 para dinero: 0.1 + 0.2 != 0.3 en coma flotante.
type Centavos int64

// TarifaIVA es la tarifa de IVA en puntos básicos: 1900 = 19 % (ADR-0007).
type TarifaIVA int64

// Tarifas permitidas (D-14).
const (
	TarifaIVA0  TarifaIVA = 0
	TarifaIVA5  TarifaIVA = 500
	TarifaIVA19 TarifaIVA = 1900
)

// NuevaTarifaIVA valida que los puntos básicos correspondan a una tarifa permitida.
func NuevaTarifaIVA(puntosBasicos int64) (TarifaIVA, error) {
	t := TarifaIVA(puntosBasicos)
	if !t.valida() {
		return 0, fmt.Errorf("%w: %d puntos básicos", ErrTarifaIVAInvalida, puntosBasicos)
	}
	return t, nil
}

func (t TarifaIVA) valida() bool {
	return t == TarifaIVA0 || t == TarifaIVA5 || t == TarifaIVA19
}

// Unidad es la unidad de medida de un ítem (D-13).
//
// Go no tiene enums como Java. El patrón idiomático es un tipo propio sobre string
// más constantes: el compilador no deja pasar un string suelto donde se espera una
// Unidad sin una conversión explícita, y ParseUnidad valida los datos externos.
type Unidad string

// Unidades permitidas (lista cerrada, D-13).
const (
	UnidadUnidad    Unidad = "und"
	UnidadKilogramo Unidad = "kg"
	UnidadGramo     Unidad = "g"
	UnidadLitro     Unidad = "l"
	UnidadMililitro Unidad = "ml"
)

// ParseUnidad convierte un texto externo (JSON, base de datos) en una Unidad válida.
func ParseUnidad(s string) (Unidad, error) {
	u := Unidad(s)
	if !u.valida() {
		return "", fmt.Errorf("%w: %q", ErrUnidadInvalida, s)
	}
	return u, nil
}

func (u Unidad) valida() bool {
	switch u {
	case UnidadUnidad, UnidadKilogramo, UnidadGramo, UnidadLitro, UnidadMililitro:
		return true
	}
	return false
}

// PermiteFracciones indica si las cantidades en esta unidad admiten decimales:
// 2,5 kg de harina sí; 2,5 pasteles no (D-13).
func (u Unidad) PermiteFracciones() bool {
	return u != UnidadUnidad
}

// TipoItem distingue lo que se vende de lo que se compra (RN-02).
type TipoItem string

// Tipos de ítem.
const (
	TipoProductoVendible TipoItem = "PRODUCTO_VENDIBLE"
	TipoMateriaPrima     TipoItem = "MATERIA_PRIMA"
)

// normalizarNombre quita espacios sobrantes y valida la longitud en caracteres
// (runas), no en bytes: "pañuelo" tiene 7 caracteres pero 8 bytes en UTF-8.
func normalizarNombre(nombre string) (string, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", ErrNombreVacio
	}
	if utf8.RuneCountInString(nombre) > LongitudMaximaNombre {
		return "", fmt.Errorf("%w (%d caracteres)", ErrNombreMuyLargo, LongitudMaximaNombre)
	}
	return nombre, nil
}
