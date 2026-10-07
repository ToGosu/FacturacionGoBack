package domain

// Categoria agrupa ítems (RF-CAT-04). La unicidad del nombre la garantiza la
// base de datos: el dominio solo no puede saber qué otras categorías existen.
type Categoria struct {
	id     CategoriaID
	nombre string
}

// NuevaCategoria crea una categoría con nombre válido.
func NuevaCategoria(id CategoriaID, nombre string) (*Categoria, error) {
	if id == "" {
		return nil, ErrIDVacio
	}
	nombre, err := normalizarNombre(nombre)
	if err != nil {
		return nil, err
	}
	return &Categoria{id: id, nombre: nombre}, nil
}

// ID devuelve el identificador de la categoría.
func (c *Categoria) ID() CategoriaID { return c.id }

// Nombre devuelve el nombre de la categoría.
func (c *Categoria) Nombre() string { return c.nombre }
