package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/cbalite/backend/internal/domain"
	"github.com/cbalite/backend/internal/middleware"
	wsHandler "github.com/cbalite/backend/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Auth handlers are now in auth_handlers.go

func (app *Application) getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	// Get user from database
	var user domain.User
	var avatar *string
	query := `
		SELECT id, email, username, first_name, last_name, avatar, is_active, is_verified, last_seen, created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true
	`
	
	err := app.DB.QueryRow(query, claims.UserID).Scan(
		&user.ID, &user.Email, &user.Username, &user.FirstName,
		&user.LastName, &avatar, &user.IsActive, &user.IsVerified,
		&user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	
	// Handle NULL avatar
	if avatar != nil {
		user.Avatar = *avatar
	}
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get current user")
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (app *Application) updateCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Update current user endpoint"})
}

func (app *Application) createTeamHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Team name is required")
		return
	}

	teamID := uuid.New().String()
	
	tx, err := app.DB.BeginTransaction(r.Context())
	if err != nil {
		app.Logger.WithError(err).Error("Failed to start transaction")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	// Create team
	_, err = tx.Exec(`
		INSERT INTO teams (id, name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, teamID, req.Name, req.Description, claims.UserID)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to create team")
		respondWithError(w, http.StatusInternalServerError, "Failed to create team")
		return
	}

	// Add owner as member
	_, err = tx.Exec(`
		INSERT INTO team_members (team_id, user_id, role, joined_at)
		VALUES ($1, $2, 'owner', NOW())
	`, teamID, claims.UserID)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to add team owner as member")
		respondWithError(w, http.StatusInternalServerError, "Failed to create team")
		return
	}

	// Create default general channel
	channelID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO channels (id, team_id, name, description, type, created_by, created_at, updated_at)
		VALUES ($1, $2, 'general', 'General discussion', 'general', $3, NOW(), NOW())
	`, channelID, teamID, claims.UserID)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to create default channel")
		respondWithError(w, http.StatusInternalServerError, "Failed to create team")
		return
	}

	if err = tx.Commit(); err != nil {
		app.Logger.WithError(err).Error("Failed to commit transaction")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	team := map[string]interface{}{
		"id":          teamID,
		"name":        req.Name,
		"description": req.Description,
		"owner_id":    claims.UserID,
	}

	respondWithJSON(w, http.StatusCreated, team)
}

func (app *Application) getTeamsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	query := `
		SELECT t.id, t.name, t.description, t.owner_id, t.created_at, t.updated_at,
		       tm.role, tm.joined_at
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name
	`
	
	rows, err := app.DB.Query(query, claims.UserID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get user teams")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var teams []map[string]interface{}
	
	for rows.Next() {
		var id, name, description, ownerID, role string
		var createdAt, updatedAt, joinedAt time.Time
		
		err := rows.Scan(
			&id, &name, &description, &ownerID,
			&createdAt, &updatedAt, &role, &joinedAt,
		)
		if err != nil {
			app.Logger.WithError(err).Error("Failed to scan team row")
			continue
		}
		
		team := map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"owner_id":    ownerID,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"role":        role,
			"joined_at":   joinedAt,
		}
		
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		app.Logger.WithError(err).Error("Error iterating team rows")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Ensure we always return an array, even if empty
	if teams == nil {
		teams = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, teams)
}

func (app *Application) getTeamHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Get team endpoint"})
}

func (app *Application) updateTeamHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Update team endpoint"})
}

func (app *Application) deleteTeamHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Delete team endpoint"})
}

func (app *Application) getTeamMembersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	// Verify user has access to this team
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)
	`, teamID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check team membership")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this team")
		return
	}

	query := `
		SELECT tm.user_id, tm.role, tm.joined_at, tm.updated_at,
		       u.email, u.username, u.first_name, u.last_name, u.avatar
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at
	`
	
	rows, err := app.DB.Query(query, teamID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get team members")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	
	for rows.Next() {
		var userID, role, email, username, firstName, lastName string
		var avatar *string
		var joinedAt, updatedAt time.Time
		
		err := rows.Scan(&userID, &role, &joinedAt, &updatedAt,
			&email, &username, &firstName, &lastName, &avatar)
		if err != nil {
			app.Logger.WithError(err).Error("Failed to scan team member row")
			continue
		}
		
		member := map[string]interface{}{
			"user_id":    userID,
			"role":       role,
			"joined_at":  joinedAt,
			"updated_at": updatedAt,
			"user": map[string]interface{}{
				"email":      email,
				"username":   username,
				"first_name": firstName,
				"last_name":  lastName,
			},
		}
		
		if avatar != nil {
			member["user"].(map[string]interface{})["avatar"] = *avatar
		}
		
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		app.Logger.WithError(err).Error("Error iterating team member rows")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Ensure we always return an array, even if empty
	if members == nil {
		members = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, members)
}

func (app *Application) inviteTeamMemberHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	var req struct {
		Email    string `json:"email"`
		Role     string `json:"role"`
		Username string `json:"username,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	// Verify that the requesting user has permission to invite members (owner or admin)
	var userRole string
	err := app.DB.QueryRow(`
		SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2
	`, teamID, claims.UserID).Scan(&userRole)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusForbidden, "Access denied to this team")
		} else {
			app.Logger.WithError(err).Error("Failed to check user role")
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if userRole != "owner" && userRole != "admin" {
		respondWithError(w, http.StatusForbidden, "Only team owners and admins can invite members")
		return
	}

	// Find user by email or username
	var userID string
	var userQuery string
	var queryParam string
	
	if req.Username != "" {
		userQuery = `SELECT id FROM users WHERE username = $1 AND is_active = true`
		queryParam = req.Username
	} else {
		userQuery = `SELECT id FROM users WHERE email = $1 AND is_active = true`
		queryParam = req.Email
	}

	err = app.DB.QueryRow(userQuery, queryParam).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			app.Logger.WithError(err).Error("Failed to find user")
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Check if user is already a member
	var existingMember bool
	err = app.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)
	`, teamID, userID).Scan(&existingMember)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check existing membership")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if existingMember {
		respondWithError(w, http.StatusConflict, "User is already a member of this team")
		return
	}

	// Add user to team
	_, err = app.DB.Exec(`
		INSERT INTO team_members (team_id, user_id, role, joined_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, teamID, userID, req.Role)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to add team member")
		respondWithError(w, http.StatusInternalServerError, "Failed to add team member")
		return
	}

	// Get user details for response
	var user struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		Username  string    `json:"username"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Avatar    *string   `json:"avatar"`
	}

	err = app.DB.QueryRow(`
		SELECT id, email, username, first_name, last_name, avatar
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Username, 
		&user.FirstName, &user.LastName, &user.Avatar)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get user details")
		// Still return success since the member was added
	}

	response := map[string]interface{}{
		"message":  "Team member added successfully",
		"user_id":  userID,
		"role":     req.Role,
		"user":     user,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (app *Application) removeTeamMemberHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Remove team member endpoint"})
}

func (app *Application) createChannelHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Create channel endpoint"})
}

func (app *Application) getChannelsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	// Verify user has access to this team
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)
	`, teamID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check team membership")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this team")
		return
	}

	query := `
		SELECT c.id, c.name, c.description, c.type, c.is_private, c.created_by, c.created_at, c.updated_at
		FROM channels c
		WHERE c.team_id = $1
		ORDER BY c.name
	`
	
	rows, err := app.DB.Query(query, teamID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get team channels")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	
	for rows.Next() {
		var id, name, description, channelType, createdBy string
		var isPrivate bool
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(&id, &name, &description, &channelType, &isPrivate, &createdBy, &createdAt, &updatedAt)
		if err != nil {
			app.Logger.WithError(err).Error("Failed to scan channel row")
			continue
		}
		
		channel := map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"type":        channelType,
			"is_private":  isPrivate,
			"created_by":  createdBy,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		}
		
		channels = append(channels, channel)
	}

	if err = rows.Err(); err != nil {
		app.Logger.WithError(err).Error("Error iterating channel rows")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Ensure we always return an array, even if empty
	if channels == nil {
		channels = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, channels)
}

func (app *Application) getChannelHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Get channel endpoint"})
}

func (app *Application) updateChannelHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Update channel endpoint"})
}

func (app *Application) deleteChannelHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Delete channel endpoint"})
}

func (app *Application) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	channelID := vars["channelId"]

	var req struct {
		Content string `json:"content"`
		Type    string `json:"type"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.Logger.WithError(err).Error("Failed to decode JSON request body")
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	if req.Type == "" {
		req.Type = "text"
	}

	// Verify user has access to this channel (through team membership)
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM channels c
			JOIN team_members tm ON c.team_id = tm.team_id
			WHERE c.id = $1 AND tm.user_id = $2
		)
	`, channelID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check channel access")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this channel")
		return
	}

	// Get the team_id for this channel
	var teamID string
	err = app.DB.QueryRow(`SELECT team_id FROM channels WHERE id = $1`, channelID).Scan(&teamID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get team_id for channel")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	messageID := uuid.New().String()

	query := `
		INSERT INTO messages (id, team_id, channel_id, user_id, content, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	
	_, err = app.DB.Exec(query, messageID, teamID, channelID, claims.UserID, req.Content, req.Type)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to create message")
		respondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Get user info for the response
	var username, firstName, lastName string
	err = app.DB.QueryRow(`
		SELECT username, first_name, last_name FROM users WHERE id = $1
	`, claims.UserID).Scan(&username, &firstName, &lastName)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get user info")
		// Continue anyway with basic info
		username = claims.Username
	}

	message := map[string]interface{}{
		"id":         messageID,
		"content":    req.Content,
		"type":       req.Type,
		"sender_id":  claims.UserID,
		"created_at": time.Now(),
		"updated_at": time.Now(),
		"sender": map[string]interface{}{
			"username":   username,
			"first_name": firstName,
			"last_name":  lastName,
		},
	}

	respondWithJSON(w, http.StatusCreated, message)
}

func (app *Application) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	channelID := vars["channelId"]

	// Verify user has access to this channel (through team membership)
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM channels c
			JOIN team_members tm ON c.team_id = tm.team_id
			WHERE c.id = $1 AND tm.user_id = $2
		)
	`, channelID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check channel access")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this channel")
		return
	}

	// Get limit from query parameter
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	query := `
		SELECT m.id, m.content, m.type, m.user_id, m.created_at, m.updated_at,
		       u.username, u.first_name, u.last_name
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`
	
	rows, err := app.DB.Query(query, channelID, limit)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get messages")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	
	for rows.Next() {
		var id, content, messageType, senderID, username, firstName, lastName string
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(&id, &content, &messageType, &senderID, &createdAt, &updatedAt,
			&username, &firstName, &lastName)
		if err != nil {
			app.Logger.WithError(err).Error("Failed to scan message row")
			continue
		}
		
		message := map[string]interface{}{
			"id":         id,
			"content":    content,
			"type":       messageType,
			"sender_id":  senderID,
			"created_at": createdAt,
			"updated_at": updatedAt,
			"sender": map[string]interface{}{
				"username":   username,
				"first_name": firstName,
				"last_name":  lastName,
			},
		}
		
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		app.Logger.WithError(err).Error("Error iterating message rows")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Reverse the order to show oldest first (since we queried DESC for limit)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Ensure we always return an array, even if empty
	if messages == nil {
		messages = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, messages)
}

func (app *Application) updateMessageHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Update message endpoint"})
}

func (app *Application) deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Delete message endpoint"})
}

func (app *Application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		AssigneeID  string `json:"assignee_id,omitempty"`
		DueDate     string `json:"due_date,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Task title is required")
		return
	}

	// Verify user has access to this team
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)
	`, teamID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check team membership")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this team")
		return
	}

	taskID := uuid.New().String()

	query := `
		INSERT INTO tasks (id, team_id, title, description, status, priority, assignee_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'todo', $5, $6, $7, NOW(), NOW())
	`
	
	var assigneeID *string
	if req.AssigneeID != "" {
		assigneeID = &req.AssigneeID
	}
	
	_, err = app.DB.Exec(query, taskID, teamID, req.Title, req.Description, req.Priority, assigneeID, claims.UserID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to create task")
		respondWithError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	task := map[string]interface{}{
		"id":          taskID,
		"title":       req.Title,
		"description": req.Description,
		"status":      "todo",
		"priority":    req.Priority,
		"created_by":  claims.UserID,
	}
	
	if assigneeID != nil {
		task["assignee_id"] = *assigneeID
	}

	respondWithJSON(w, http.StatusCreated, task)
}

func (app *Application) getTasksHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	// Verify user has access to this team
	var memberExists bool
	err := app.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)
	`, teamID, claims.UserID).Scan(&memberExists)
	
	if err != nil {
		app.Logger.WithError(err).Error("Failed to check team membership")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !memberExists {
		respondWithError(w, http.StatusForbidden, "Access denied to this team")
		return
	}

	query := `
		SELECT t.id, t.title, t.description, t.status, t.priority, 
		       t.assignee_id, t.due_date, t.created_by, t.created_at, t.updated_at
		FROM tasks t
		WHERE t.team_id = $1
		ORDER BY t.created_at DESC
	`
	
	rows, err := app.DB.Query(query, teamID)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to get team tasks")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	
	for rows.Next() {
		var id, title, description, status, priority, createdBy string
		var assigneeID *string
		var dueDate *time.Time
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(&id, &title, &description, &status, &priority, 
			&assigneeID, &dueDate, &createdBy, &createdAt, &updatedAt)
		if err != nil {
			app.Logger.WithError(err).Error("Failed to scan task row")
			continue
		}
		
		task := map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"status":      status,
			"priority":    priority,
			"created_by":  createdBy,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		}
		
		if assigneeID != nil {
			task["assignee_id"] = *assigneeID
		}
		
		if dueDate != nil {
			task["due_date"] = *dueDate
		}
		
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		app.Logger.WithError(err).Error("Error iterating task rows")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Ensure we always return an array, even if empty
	if tasks == nil {
		tasks = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (app *Application) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Get task endpoint"})
}

func (app *Application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Update task endpoint"})
}

func (app *Application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Delete task endpoint"})
}

func (app *Application) createTaskCommentHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Create task comment endpoint"})
}

func (app *Application) getTaskCommentsHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{"message": "Get task comments endpoint"})
}

func (app *Application) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Try to get token from query params or headers
	var userID, teamID string = "anonymous", ""
	
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	if token != "" {
		// Validate token and get user info
		if claims, err := app.AuthMiddleware.ValidateToken(token); err == nil {
			userID = claims.UserID
			
			// Get user's team (for now, just use first team they're a member of)
			var teamIDFromDB string
			err := app.DB.QueryRow(`
				SELECT team_id FROM team_members 
				WHERE user_id = $1 
				LIMIT 1
			`, claims.UserID).Scan(&teamIDFromDB)
			if err == nil {
				teamID = teamIDFromDB
			}
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.Logger.WithError(err).Error("Failed to upgrade connection")
		return
	}

	clientID := uuid.New().String()
	client := &wsHandler.Client{
		ID:     clientID,
		UserID: userID,
		TeamID: teamID,
		Conn:   conn,
		Hub:    app.WSHub,
		Send:   make(chan []byte, 256),
		Rooms:  make(map[string]bool),
	}

	app.Logger.Infof("WebSocket client connected: %s (User: %s, Team: %s)", clientID, userID, teamID)

	app.WSHub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}