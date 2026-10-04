package main

import (
	"context"
	"log"

	"github.com/GinciuuKitakaze/users/internal/app"
)

func main() {
	ctx := context.Background()

	// Инициализируем приложение.
	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем приложение.
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}

	// Корректно завершаем работу приложения.
	application.Shutdown()
}
