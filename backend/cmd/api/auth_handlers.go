package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/cbalite/backend/internal/domain"
)

func (app *Application) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	if len(req.Password) < 8 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Check if user already exists
	var exists bool
	err := app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", req.Email, req.Username).Scan(&exists)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check if user exists")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if exists {
		respondWithError(w, http.StatusConflict, "User with this email or username already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to hash password")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
		IsVerified:   false,
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, is_active, is_verified, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err = app.DB.Exec(query, user.ID, user.Email, user.Username, user.PasswordHash, 
		user.FirstName, user.LastName, user.IsActive, user.IsVerified, 
		user.LastSeen, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to create user")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Generate tokens
	accessToken, err := app.AuthMiddleware.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to generate access token")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	refreshToken, err := app.AuthMiddleware.GenerateRefreshToken(user.ID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to generate refresh token")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Remove sensitive data
	user.PasswordHash = ""

	response := map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EmailOrUsername == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Missing email/username or password")
		return
	}

	// Find user by email or username
	var user domain.User
	var avatar *string
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, avatar, is_active, is_verified, last_seen, created_at, updated_at
		FROM users 
		WHERE (email = $1 OR username = $1) AND is_active = true
	`
	
	err := app.DB.QueryRow(query, req.EmailOrUsername).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &avatar, &user.IsActive,
		&user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	
	// Handle NULL avatar
	if avatar != nil {
		user.Avatar = *avatar
	}
	if err != nil {
		app.Logger.WithError(err).Debug("User not found")
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		app.Logger.WithError(err).Debug("Invalid password")
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Update last seen
	_, err = app.DB.Exec("UPDATE users SET last_seen = $1 WHERE id = $2", time.Now(), user.ID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to update last seen")
		// Continue anyway
	}

	// Generate tokens
	accessToken, err := app.AuthMiddleware.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to generate access token")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	refreshToken, err := app.AuthMiddleware.GenerateRefreshToken(user.ID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to generate refresh token")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Remove sensitive data
	user.PasswordHash = ""

	response := map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (app *Application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate refresh token
	claims, err := app.AuthMiddleware.ValidateToken(req.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Get user
	var user domain.User
	var avatar *string
	query := `
		SELECT id, email, username, first_name, last_name, avatar, is_active, is_verified
		FROM users 
		WHERE id = $1 AND is_active = true
	`
	
	err = app.DB.QueryRow(query, claims.UserID).Scan(
		&user.ID, &user.Email, &user.Username, &user.FirstName,
		&user.LastName, &avatar, &user.IsActive, &user.IsVerified,
	)
	
	// Handle NULL avatar
	if avatar != nil {
		user.Avatar = *avatar
	}
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Generate new access token
	accessToken, err := app.AuthMiddleware.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to generate access token")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := map[string]interface{}{
		"access_token": accessToken,
		"user":         user,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	// For now, just return success
	// In a full implementation, you might want to blacklist the token
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}