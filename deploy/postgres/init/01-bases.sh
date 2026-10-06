#!/bin/sh
# Crea una base de datos y un usuario dueño por servicio (ADR-0006).
# Postgres ejecuta este script una sola vez, al inicializar un volumen vacío.
set -eu

crear_base() {
	base="$1"
	clave="$2"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
		-v base="$base" -v clave="$clave" <<-'SQL'
		CREATE USER :"base" WITH PASSWORD :'clave';
		CREATE DATABASE :"base" OWNER :"base";
		REVOKE ALL ON DATABASE :"base" FROM PUBLIC;
	SQL
}

crear_base identidad "$DB_PASSWORD_IDENTIDAD"
crear_base productos_inventario "$DB_PASSWORD_PRODUCTOS_INVENTARIO"
crear_base ventas "$DB_PASSWORD_VENTAS"
crear_base notificaciones "$DB_PASSWORD_NOTIFICACIONES"
crear_base reportes "$DB_PASSWORD_REPORTES"
