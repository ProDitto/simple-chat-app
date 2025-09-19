package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"chat-app/internal/chat"
	"chat-app/internal/domain/user"
	"chat-app/internal/handler"
)

func main() {
	// --- Dependencies ---
	// In a real app, you might get this from a config file or env vars
	inMemoryUsers := map[string]string{
		"alice":   "password123",
		"bob":     "password123",
		"charlie": "password123",
		"dave":    "password123",
	}
	userService := user.NewUserService(inMemoryUsers)
	hub := chat.NewHub()
	go hub.Run()
	chatHandler := handler.NewHandler(hub, userService)

	// --- Router Setup ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- Routes ---
	// Serve frontend files
	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)

	// API routes
	r.Post("/login", chatHandler.Login)
	r.Get("/ws", chatHandler.ServeWs)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
