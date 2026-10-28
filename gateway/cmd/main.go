package main

import (
	"context"
	"log"
	"time"

	"github.com/GinciuuKitakaze/gateway/internal/app"
)

func main() {
	// Инициализируем приложение.
	application, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем приложение и ожидаем сигнал завершения.
	runErr := application.Run()

	// Даём приложению время корректно завершить
	// текущие HTTP-запросы и закрыть соединения.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		log.Printf("shutdown application: %v", err)
	}

	if runErr != nil {
		log.Printf("run application: %v", runErr)
	}
}
