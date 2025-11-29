package main

import (
	"log/slog"
	"os"

	"github.com/duykhoa/gopass/internal/pico"
)

func main() {
	// Configure slog
	file, err := os.OpenFile("logs/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("Can't create log file")
	}

	defer file.Close()

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum logging level to DEBUG
		// Optionally, add source file and line number:
		AddSource: false,
	})

	// 3. Create a new logger with the file handler.
	logger := slog.New(handler)

	// 4. Set the new logger as the default for the application.
	slog.SetDefault(logger)

	application := pico.NewPico()
	application.Run()
}
