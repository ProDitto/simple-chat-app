package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"chat-app/internal/chat"
	"chat-app/internal/domain/user"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for this MVP
		return true
	},
}

type Handler struct {
	hub         *chat.Hub
	userService *user.UserService
}

func NewHandler(h *chat.Hub, us *user.UserService) *Handler {
	return &Handler{
		hub:         h,
		userService: us,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Login handles user authentication.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if h.userService.Authenticate(req.Username, req.Password) {
		json.NewEncoder(w).Encode(loginResponse{Success: true, Message: "Login successful"})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(loginResponse{Success: false, Message: "Invalid credentials"})
	}
}

// ServeWs handles websocket requests from the peer.
func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("Username not provided")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Check if user is already connected
	if _, ok := h.hub.Clients[username]; ok {
		log.Printf("User %s already connected", username)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "User already connected"))
		conn.Close()
		return
	}

	client := chat.NewClient(h.hub, conn, username)
	h.hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
