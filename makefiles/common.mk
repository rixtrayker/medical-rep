APP_NAME := myapp
CMD_DIR := cmd
ENTRY := $(CMD_DIR)/main.go
BIN := bin/$(APP_NAME)

DB_URL ?= postgres://user:pass@localhost:5432/$(APP_NAME)?sslmode=disable
MIGRATE := migrate -path db/migrations -database $(DB_URL)

ENV_FILE := .env
