# ADR-0006 — PostgreSQL por servicio, `sqlc` + `pgx` y migraciones

**Estado:** Aceptado (2026-10-06)

**Contexto.** Cada servicio es dueño de su dato. Venir de Spring/JPA hace tentador usar un ORM.

**Decisión.** Una **instancia** de PostgreSQL con **una base y un usuario por servicio** (aislamiento lógico, bajo costo). Acceso con **`sqlc` + `pgx`** (SQL escrito a mano, código Go tipado generado) y migraciones con `golang-migrate`.

**Consecuencias.**
- (+) Obliga a pensar en SQL, que se pregunta en entrevistas, y encaja con los repositorios hexagonales.
- (+) Sin "magia" de ORM; fácil de depurar.
- (−) Más SQL manual que con GORM.
- (−) Una sola instancia es un punto único de falla; aceptable para este alcance y se documenta.

**Alternativas.** GORM (más familiar viniendo de JPA, pero oculta el SQL) · una instancia por servicio (más fiel al ideal, más consumo de recursos).
