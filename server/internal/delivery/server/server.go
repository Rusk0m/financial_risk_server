package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server представляет HTTP сервер с настройками
type Server struct {
	httpServer *http.Server
}

// NewServer создаёт новый HTTP сервер
func NewServer(cfg *ServerConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		},
	}
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port         string
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
	IdleTimeout  int // seconds
}

// Run запускает сервер
func (s *Server) Run() error {
	log.Printf("🚀 HTTP-сервер запущен на порту %s", s.httpServer.Addr)
	log.Println("   Нажмите Ctrl+C для остановки")

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("🛑 Получен сигнал завершения, останавливаем сервер...")

		// Даём 5 секунд на завершение активных запросов
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("❌ Ошибка при остановке сервера: %v", err)
		}
		close(idleConnsClosed)
	}()

	// Запускаем сервер
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	log.Println("✅ Сервер остановлен корректно")
	return nil
}
