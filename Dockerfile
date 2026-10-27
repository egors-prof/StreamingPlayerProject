# syntax=docker/dockerfile:1
# Один Dockerfile для всех Go-сервисов. Какой сервис собирать, задаёт
# аргумент SERVICE (имя папки в services/). Контекст сборки — корень
# репозитория, потому что сервисам нужен общий модуль pkg/.
#
#   docker build --build-arg SERVICE=auth .

FROM golang:1.25-alpine AS build
ARG SERVICE
WORKDIR /src

# Сначала копируем только go.mod/go.sum и скачиваем зависимости. Docker
# кеширует этот слой, и при изменении кода зависимости не качаются заново.
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum ./services/${SERVICE}/
RUN cd services/${SERVICE} && go mod download

COPY pkg ./pkg
COPY services/${SERVICE} ./services/${SERVICE}
# CGO_ENABLED=0 — статический бинарник, которому не нужна libc в образе.
RUN cd services/${SERVICE} && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd

# Финальный образ distroless: только бинарник, без shell и пакетного
# менеджера. Образ весит 15–35 МБ (у auth и likes внутри Swagger UI)
# вместо сотен мегабайт у golang-образа, и меньше поверхность атаки.
# Пользователь nonroot — процесс работает не от root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
