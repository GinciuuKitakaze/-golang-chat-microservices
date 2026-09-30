# ─── Настройки приложения ──────────────────────────────────────────────
# Если переменная APP_NAME не задана извне, используем "account"
APP_NAME ?= users

# Имя исполняемого файла (будет лежать в папке bin/)
BINARY_NAME ?= bin/$(APP_NAME)

# Путь к главному пакету (где лежит main.go)
MAIN_PATH ?= ./cmd

# Версия: берётся из git-тега, если есть; иначе "dev"
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Время сборки в формате UTC (используется для инжекта в бинарник)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Флаги линковки: передаём версию и время в переменные Go-программы (main.Version, main.BuildTime)
LDFLAGS ?= -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# ─── Список Go-файлов для форматирования и линтера ──────────────────
# Находим все .go файлы, исключая vendor и .git
GOFILES ?= $(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*")

# ─── Цель по умолчанию ────────────────────────────────────────────────
# Если ввести просто "make", будет выполнена цель build
.DEFAULT_GOAL := build

# ─── Сборка бинарника ──────────────────────────────────────────────────
build:
	@echo "Building $(APP_NAME) ..."      # выводим сообщение
	@mkdir -p bin                          # создаём папку bin, если её нет
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)   # компилируем

# ─── Запуск (сборка + запуск) ────────────────────────────────────────
run: build                                # сначала выполняем build, затем run
	@echo "Running $(APP_NAME) ..."
	./$(BINARY_NAME)                       # запускаем собранный бинарник

# ─── Запуск без пересборки ───────────────────────────────────────────
run-only:                                 # предполагаем, что бинарник уже есть
	@echo "Running $(APP_NAME) ..."
	./$(BINARY_NAME)

# ─── Установка зависимостей ──────────────────────────────────────────
deps:
	@echo "Installing dependencies ..."
	go mod download                        # скачивает модули из go.mod
	go mod tidy                            # очищает go.mod и go.sum

# ─── Обновление зависимостей ─────────────────────────────────────────
deps-update:
	@echo "Updating dependencies ..."
	go get -u ./...                        # обновляет все зависимости (рекурсивно)
	go mod tidy

# ─── Форматирование кода ──────────────────────────────────────────────
fmt:
	@echo "Formatting code ..."
	gofmt -s -w $(GOFILES)                 # стандартное форматирование
	# gofumpt -l -w -s $(GOFILES)          # (опционально) улучшенный форматтер
	# goimports -l -w $(GOFILES)           # (опционально) сортировка импортов

# ─── Линтер (проверка качества кода) ────────────────────────────────
lint:                                     # требует установленного golangci-lint
	@echo "Running linter ..."
	golangci-lint run --timeout=10m -v ./...

# ─── Запуск тестов ────────────────────────────────────────────────────
test:
	@echo "Running tests ..."
	go test -v ./...

# ─── Тесты с покрытием ───────────────────────────────────────────────
test-coverage:
	@echo "Running tests with coverage ..."
	go test -v -coverprofile=coverage.out ./...   # записываем профиль покрытия
	go tool cover -html=coverage.out -o coverage.html   # генерируем HTML-отчёт
	@echo "Coverage report generated: coverage.html"

# ─── Бенчмарки (замер производительности) ──────────────────────────
bench:
	@echo "Running benchmarks ..."
	go test -bench=. ./...

# ─── Очистка артефактов сборки ──────────────────────────────────────
clean:
	@echo "Cleaning build artifacts ..."
	rm -rf bin/                           # удаляем папку с бинарниками
	rm -f coverage.out coverage.html      # удаляем файлы покрытия

# ─── Установка приложения в GOPATH/bin ─────────────────────────────
install: build
	@echo "Installing $(APP_NAME) ..."
	go install $(LDFLAGS) $(MAIN_PATH)    # аналог go install с флагами

# ─── Генерация документации (godoc) ────────────────────────────────
docs:
	@echo "Generating documentation ..."
	godoc -http=:6060                     # запускает локальный сервер документации

# ─── Генерация документации (swagger) ────────────────────────────────
swagger-gen:
	@swag init -g cmd/main.go --parseDependency --parseInternal
	@echo "Swagger docs regenerate"

# ─── Справка ──────────────────────────────────────────────────────────
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  run-only       - Run the application (without building)"
	@echo "  deps           - Install dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install the application"
	@echo "  docs           - Generate documentation"
	@echo "  help           - Show this help"

# ─── Комбинированные цели для разработки ────────────────────────────
dev: fmt lint test build                 # форматирование, линтер, тесты, сборка

ci: deps fmt lint test build             # то же, но с загрузкой зависимостей

# ─── Объявляем цели как "фейковые" (не файлы) ──────────────────────
.PHONY: build run run-only deps deps-update fmt lint test test-coverage bench clean install docs help dev ci