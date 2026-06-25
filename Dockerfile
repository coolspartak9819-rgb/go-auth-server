# === Этап 1: Сборка бинарника ===
FROM golang:1.26-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем кэш модулей
RUN go mod download

# Копируем весь исходный код проекта
COPY . .

# Компилируем приложение в один статически связанный бинарный файл "server"
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# === Этап 2: Финальный легковесный образ ===
FROM alpine:latest

WORKDIR /root/

# Копируем скомпилированный бинарник из предыдущего этапа
COPY --from=builder /app/server .

# Открываем порт наружу
EXPOSE 8082

# Запускаем наше приложение
CMD ["./server"]