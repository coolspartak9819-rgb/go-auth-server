package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myapi/config"
	"myapi/database"
	"myapi/handlers"
	"myapi/middleware"
	"myapi/queue"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Инфраструктура
	database.InitRedis(ctx)
	queue.InitKafka()
	queue.StartSecurityWorker(ctx)

	// 2. Маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/charge", handlers.ChargeHandler)
	mux.HandleFunc("/buy-premium", handlers.BuyPremiumHandler) // 🔥 ДОБАВИЛИ РОУТ САГИ
	mux.Handle("/metrics", promhttp.Handler())

	protectedMux := middleware.SecurityAndIdempotencyMiddleware(mux)
	finalMux := middleware.RateLimitMiddleware(protectedMux)

	server := &http.Server{
		Addr:    config.ListenAddr,
		Handler: finalMux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("🚀 Сервер запущен в Docker на порту %s...\n", config.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("🔥 Ошибка запуска сервера: %v", err)
		}
	}()

	<-stop
	fmt.Println("\n⏳ Начинается процесс мягкой остановки сервера (Graceful Shutdown)...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Ошибка при остановке HTTP-сервера: %v", err)
	}

	cancel()
	queue.CloseKafka()
	fmt.Println("💚 Сервер успешно и безопасно остановлен.")
}
