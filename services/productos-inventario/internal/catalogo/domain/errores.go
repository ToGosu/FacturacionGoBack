// Package domain contiene las entidades y reglas de negocio del catálogo.
// No importa nada de infraestructura (HTTP, SQL, NATS): se prueba solo con Go.
package domain

import "errors"

// Errores de negocio del catálogo.
//
// En Go los errores son valores que se devuelven, no excepciones que se lanzan.
// Estos "errores centinela" cumplen el papel de las excepciones de negocio de Java:
// las capas externas los reconocen con errors.Is(err, domain.ErrPrecioInvalido)
// aunque vengan envueltos con más contexto (fmt.Errorf("...: %w", err)).
var (
	ErrIDVacio               = errors.New("el identificador es obligatorio")
	ErrNombreVacio           = errors.New("el nombre es obligatorio")
	ErrNombreMuyLargo        = errors.New("el nombre supera la longitud máxima")
	ErrUnidadInvalida        = errors.New("unidad de medida no permitida")
	ErrTarifaIVAInvalida     = errors.New("tarifa de IVA no permitida: use 0 %, 5 % o 19 %")
	ErrPrecioInvalido        = errors.New("el precio debe ser mayor que cero")
	ErrPrecioSinCambio       = errors.New("el precio nuevo es igual al vigente")
	ErrVigenciaInvalida      = errors.New("el precio nuevo no puede empezar antes que el vigente")
	ErrItemInactivo          = errors.New("el ítem ya está inactivo")
	ErrNoAplicaAMateriaPrima = errors.New("la operación solo aplica a productos vendibles")
)
