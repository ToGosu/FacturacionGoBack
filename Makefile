# Tareas comunes del monorepo.
#
# Cada servicio es un módulo Go independiente (ADR-0010). `go build ./...` no cruza
# módulos desde la raíz del go.work, así que se genera un objetivo por módulo
# (build.services/ventas, test.pkg, ...) y se usa `go -C <dir>` para no depender
# de `cd` ni de bucles de shell: funciona igual en Windows (cmd) y en Linux.

MODULES := $(patsubst %/go.mod,%,$(wildcard pkg/go.mod services/*/go.mod))

GO_IMAGE := golang:1.27
COMPOSE  := docker compose -f deploy/docker-compose.yml $(if $(wildcard .env),--env-file .env)

# Variables opcionales: make test RACE=1
TEST_FLAGS := $(if $(RACE),-race)

.DEFAULT_GOAL := help

.PHONY: help build vet test test-race lint fmt tidy run up down logs ps

help: ## Muestra esta ayuda
	@echo Objetivos: build vet test test-race lint fmt tidy run SERVICE=x up down logs ps
	@echo Modulos detectados: $(if $(MODULES),$(MODULES),ninguno)

build: $(addprefix build.,$(MODULES)) ## Compila todos los módulos
vet: $(addprefix vet.,$(MODULES)) ## go vet en todos los módulos
test: $(addprefix test.,$(MODULES)) ## Pruebas de todos los módulos
lint: $(addprefix lint.,$(MODULES)) ## golangci-lint (incluye gofmt/goimports)
fmt: $(addprefix fmt.,$(MODULES)) ## Formatea el código con golangci-lint fmt
tidy: $(addprefix tidy.,$(MODULES)) ## go mod tidy en todos los módulos

$(addprefix build.,$(MODULES)): build.%:
	go -C $* build ./...

$(addprefix vet.,$(MODULES)): vet.%:
	go -C $* vet ./...

$(addprefix test.,$(MODULES)): test.%:
	go -C $* test $(TEST_FLAGS) ./...

$(addprefix lint.,$(MODULES)): lint.%:
	cd $* && golangci-lint run ./...

$(addprefix fmt.,$(MODULES)): fmt.%:
	cd $* && golangci-lint fmt ./...

$(addprefix tidy.,$(MODULES)): tidy.%:
	go -C $* mod tidy

# El detector de carreras (-race) necesita cgo y un compilador de C. En Windows sin gcc
# no está disponible, así que se ejecuta dentro de un contenedor Linux con Go.
test-race: ## Pruebas con -race dentro de Docker
	docker run --rm -v "$(CURDIR):/src" -v panaderia-gomod:/go/pkg/mod -w /src $(GO_IMAGE) make test RACE=1

run: ## Levanta un servicio: make run SERVICE=productos-inventario
	go -C services/$(SERVICE) run ./cmd/$(SERVICE)

up: ## Levanta la infraestructura (Postgres, NATS, Mailpit)
	$(COMPOSE) up -d

down: ## Detiene la infraestructura (los volúmenes se conservan)
	$(COMPOSE) down

logs: ## Sigue los logs de la infraestructura
	$(COMPOSE) logs -f

ps: ## Estado de los contenedores
	$(COMPOSE) ps
