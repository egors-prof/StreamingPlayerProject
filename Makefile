GO_MODULES := back/pkg back/auth_service back/StreamMain back/likes back/searchandfetch

.PHONY: up down logs test vet swagger tracks

## up: собрать и запустить всё (http://localhost:8080)
up:
	@test -f .env || (cp .env.example .env && echo "Создан .env из .env.example — поменяйте секреты")
	docker compose up --build -d
	@echo "Открой http://localhost:8080"

## down: остановить контейнеры (данные БД сохраняются в volume)
down:
	docker compose down

## logs: логи всех сервисов
logs:
	docker compose logs -f auth stream likes catalog gateway

## test: тесты всех Go-модулей с детектором гонок
test:
	@for m in $(GO_MODULES); do echo "== $$m"; (cd $$m && go test -race ./...) || exit 1; done

## vet: статический анализ
vet:
	@for m in $(GO_MODULES); do (cd $$m && go vet ./...) || exit 1; done

## swagger: перегенерировать Swagger-документацию (нужен swag: go install github.com/swaggo/swag/cmd/swag@v1.16.6)
swagger:
	cd back/auth_service && swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	cd back/likes && swag init -g cmd/main.go -o docs --parseDependency --parseInternal

## tracks: заново сгенерировать демо-треки (нужны numpy, scipy, ffmpeg)
tracks:
	python3 scripts/generate_demo_tracks.py
