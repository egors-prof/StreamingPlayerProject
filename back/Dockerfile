# syntax=docker/dockerfile:1
# Один Dockerfile для всех Go-сервисов. Какой сервис собирать, задаёт
# аргумент SERVICE (имя папки в back/). Контекст сборки — папка back/,
# потому что сервисам нужен общий модуль back/pkg.
#
#   docker build -f back/Dockerfile --build-arg SERVICE=auth_service back

FROM golang:1.25-alpine AS build
ARG SERVICE
WORKDIR /src

# Сначала копируем только go.mod/go.sum и скачиваем зависимости. Docker
# кеширует этот слой, и при изменении кода зависимости не качаются заново.
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY ${SERVICE}/go.mod ${SERVICE}/go.sum ./${SERVICE}/
RUN cd ${SERVICE} && go mod download

COPY pkg ./pkg
COPY ${SERVICE} ./${SERVICE}
# CGO_ENABLED=0 — статический бинарник, которому не нужна libc в образе.
RUN cd ${SERVICE} && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd

# Финальный образ distroless: только бинарник, без shell и пакетного
# менеджера. Образ весит 15–35 МБ (у auth и likes внутри Swagger UI)
# вместо сотен мегабайт у golang-образа, и меньше поверхность атаки.
# Пользователь nonroot — процесс работает не от root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
