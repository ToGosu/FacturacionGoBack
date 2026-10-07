package domain

import (
	"fmt"
	"time"
)

// ItemID identifica un ítem. Los IDs (UUID) los genera la capa de aplicación,
// así el dominio no depende de ninguna librería para crearlos.
type ItemID string

// CategoriaID identifica una categoría.
type CategoriaID string

// Precio es un registro del historial de precios (RN-04, RF-CAT-02).
// El valor ya incluye el IVA (D-03). VigenteHasta es nil mientras el precio rige.
type Precio struct {
	Valor        Centavos
	VigenteDesde time.Time
	VigenteHasta *time.Time
}

// Item es un producto vendible o una materia prima.
//
// Los campos están en minúscula: en Go eso los hace privados al paquete
// (equivale a private en Java). Fuera de domain solo se leen con los métodos
// de acceso y solo se modifican a través de métodos que validan las reglas.
type Item struct {
	id          ItemID
	nombre      string
	tipo        TipoItem
	unidad      Unidad
	categoriaID *CategoriaID
	tarifaIVA   TarifaIVA
	precio      *Precio // precio vigente; nil en materia prima (D-15)
	activo      bool
	creadoEn    time.Time
}

// NuevoProductoVendible crea un producto vendible: unidad fija "und" (D-13),
// tarifa de IVA permitida (D-14) y precio inicial mayor que cero.
//
// Recibe "ahora" como parámetro en vez de llamar a time.Now(): el dominio no lee
// el reloj, y las pruebas pueden fijar la hora (como inyectar un Clock en Java).
func NuevoProductoVendible(id ItemID, nombre string, tarifa TarifaIVA, precio Centavos, categoriaID *CategoriaID, ahora time.Time) (*Item, error) {
	if !tarifa.valida() {
		return nil, fmt.Errorf("%w: %d puntos básicos", ErrTarifaIVAInvalida, tarifa)
	}
	if precio <= 0 {
		return nil, ErrPrecioInvalido
	}
	it, err := nuevoItem(id, nombre, TipoProductoVendible, UnidadUnidad, categoriaID, ahora)
	if err != nil {
		return nil, err
	}
	it.tarifaIVA = tarifa
	it.precio = &Precio{Valor: precio, VigenteDesde: it.creadoEn}
	return it, nil
}

// NuevaMateriaPrima crea una materia prima: sin precio ni IVA (D-15) y con
// cualquier unidad permitida, incluidas las que admiten decimales (D-13).
func NuevaMateriaPrima(id ItemID, nombre string, unidad Unidad, categoriaID *CategoriaID, ahora time.Time) (*Item, error) {
	if !unidad.valida() {
		return nil, fmt.Errorf("%w: %q", ErrUnidadInvalida, unidad)
	}
	return nuevoItem(id, nombre, TipoMateriaPrima, unidad, categoriaID, ahora)
}

func nuevoItem(id ItemID, nombre string, tipo TipoItem, unidad Unidad, categoriaID *CategoriaID, ahora time.Time) (*Item, error) {
	if id == "" {
		return nil, ErrIDVacio
	}
	nombre, err := normalizarNombre(nombre)
	if err != nil {
		return nil, err
	}
	return &Item{
		id:          id,
		nombre:      nombre,
		tipo:        tipo,
		unidad:      unidad,
		categoriaID: categoriaID,
		activo:      true,
		creadoEn:    ahora.UTC(), // timestamps en UTC (RNF-08)
	}, nil
}

// Métodos de acceso. En Go no se usa el prefijo "Get": se escribe Nombre(), no GetNombre().

// ID devuelve el identificador del ítem.
func (it *Item) ID() ItemID { return it.id }

// Nombre devuelve el nombre del ítem.
func (it *Item) Nombre() string { return it.nombre }

// Tipo devuelve si el ítem es producto vendible o materia prima.
func (it *Item) Tipo() TipoItem { return it.tipo }

// Unidad devuelve la unidad de medida.
func (it *Item) Unidad() Unidad { return it.unidad }

// CategoriaID devuelve la categoría, o nil si no tiene.
func (it *Item) CategoriaID() *CategoriaID { return it.categoriaID }

// TarifaIVA devuelve la tarifa de IVA (0 en materia prima).
func (it *Item) TarifaIVA() TarifaIVA { return it.tarifaIVA }

// Activo indica si el ítem puede usarse en operaciones nuevas.
func (it *Item) Activo() bool { return it.activo }

// CreadoEn devuelve la fecha de creación en UTC.
func (it *Item) CreadoEn() time.Time { return it.creadoEn }

// PrecioVigente devuelve una copia del precio vigente. El segundo valor es false
// si el ítem no tiene precio (materia prima): es el idioma "valor, ok" de Go,
// parecido a devolver un Optional en Java.
func (it *Item) PrecioVigente() (Precio, bool) {
	if it.precio == nil {
		return Precio{}, false
	}
	return *it.precio, true
}

// Renombrar cambia el nombre del ítem.
func (it *Item) Renombrar(nombre string) error {
	nombre, err := normalizarNombre(nombre)
	if err != nil {
		return err
	}
	it.nombre = nombre
	return nil
}

// AsignarCategoria cambia la categoría; nil la quita.
func (it *Item) AsignarCategoria(categoriaID *CategoriaID) {
	it.categoriaID = categoriaID
}

// CambiarTarifaIVA cambia la tarifa de un producto vendible.
func (it *Item) CambiarTarifaIVA(tarifa TarifaIVA) error {
	if it.tipo != TipoProductoVendible {
		return ErrNoAplicaAMateriaPrima
	}
	if !tarifa.valida() {
		return fmt.Errorf("%w: %d puntos básicos", ErrTarifaIVAInvalida, tarifa)
	}
	it.tarifaIVA = tarifa
	return nil
}

// CambiarPrecio registra un precio nuevo que rige desde "ahora" (D-16) y cierra
// el anterior. Devuelve el precio cerrado para que la capa de persistencia lo
// guarde en el historial; las facturas pasadas no cambian (RN-04).
func (it *Item) CambiarPrecio(valor Centavos, ahora time.Time) (cerrado Precio, err error) {
	if it.tipo != TipoProductoVendible {
		return Precio{}, ErrNoAplicaAMateriaPrima
	}
	if valor <= 0 {
		return Precio{}, ErrPrecioInvalido
	}
	if valor == it.precio.Valor {
		return Precio{}, ErrPrecioSinCambio
	}
	ahora = ahora.UTC()
	if ahora.Before(it.precio.VigenteDesde) {
		return Precio{}, ErrVigenciaInvalida
	}

	cerrado = *it.precio
	cerrado.VigenteHasta = &ahora
	it.precio = &Precio{Valor: valor, VigenteDesde: ahora}
	return cerrado, nil
}

// Desactivar marca el ítem como inactivo sin borrarlo, para no romper las
// facturas históricas que lo referencian (RF-CAT-03).
func (it *Item) Desactivar() error {
	if !it.activo {
		return ErrItemInactivo
	}
	it.activo = false
	return nil
}
