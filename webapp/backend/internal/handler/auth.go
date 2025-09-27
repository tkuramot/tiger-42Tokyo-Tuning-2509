package handler

import (
	"errors"
	"log"
	"net/http"

	"backend/internal/model"
	"backend/internal/service"
	"github.com/goccy/go-json"
)

type AuthHandler struct {
	AuthSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthSvc: authSvc}
}

// ログイン時にセッションを発行し、Cookieにセットする
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("-> Received request for /api/login")

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionID, expiresAt, err := h.AuthSvc.Login(r.Context(), req.UserName, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}
