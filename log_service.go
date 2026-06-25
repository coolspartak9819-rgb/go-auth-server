package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Описываем структуру gRPC сервера
type LogServer struct{}

// Наш gRPC метод, который будут вызывать другие микросервисы
func (s *LogServer) SaveLog(ctx context.Context, req *LogRequest) (*LogResponse, error) {
	if req.Login == "" || req.Action == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Логин и действие не могут быть пустыми")
	}

	// Записываем лог в ПЕРВУЮ (основную) базу данных
	_, err := mainDB.Exec("INSERT INTO users_log (login, action) VALUES ($1, $2)", req.Login, req.Action)
	if err != nil {
		log.Printf("❌ [gRPC СЕРВЕР ЛОГОВ] Ошибка записи в БД: %v", err)
		return nil, status.Errorf(codes.Internal, "Ошибка базы данных")
	}

	fmt.Printf("📥 [gRPC СЕРВЕР ЛОГОВ] Успешно сохранил лог для: %s (%s)\n", req.Login, req.Action)
	return &LogResponse{Success: true}, nil
}

// Простые структуры-заглушки, заменяющие тяжелый Protobuf
type LogRequest struct {
	Login  string
	Action string
}

type LogResponse struct {
	Success bool
}

var mainDB *sql.DB

// Функция запуска gRPC сервера в фоне
func StartGRPCServer() {
	var err error
	connStr := "host=localhost port=5432 user=myuser password=mypassword dbname=mydb sslmode=disable"
	mainDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения gRPC к БД: %v", err)
	}

	// Слушаем внутренний порт 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("gRPC не смог занять порт 50051: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Регистрируем наш сервис логов
	// В реальном protobuf тут вызывается автосгенерированный метод, мы симулируем его регистрацию
	_ = grpcServer

	fmt.Println("🛰️ [gRPC СЕРВЕР ЛОГОВ] Запущен на порту :50051")

	go func() {
		// Кастомный роутинг gRPC запросов для демонстрации без кодогенерации
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			_ = conn // Для упрощения примера, логика вызова эмулируется клиентом напрямую
		}
	}()
}

// Прямой вызов gRPC метода внутри приложения (вместо сложного сетевого маршалинга protobuf)
func CallSaveLogGRPC(login, action string) (bool, error) {
	server := &LogServer{}
	res, err := server.SaveLog(context.Background(), &LogRequest{Login: login, Action: action})
	if err != nil {
		return false, err
	}
	return res.Success, nil
}
