package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/cbalite/backend/pkg/logger"
)

type Hub struct {
	clients    map[string]*Client
	rooms      map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	logger     *logger.Logger
	mu         sync.RWMutex
}

type Client struct {
	ID       string
	UserID   string
	TeamID   string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	Rooms    map[string]bool
}

type Message struct {
	Type      string      `json:"type"`
	Room      string      `json:"room,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type MessageType string

const (
	MessageTypeChat         MessageType = "chat"
	MessageTypeTaskUpdate   MessageType = "task_update"
	MessageTypeUserStatus   MessageType = "user_status"
	MessageTypeNotification MessageType = "notification"
	MessageTypeTyping       MessageType = "typing"
	MessageTypePresence     MessageType = "presence"
)

func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.logger.Infof("Client registered: %s (User: %s)", client.ID, client.UserID)

	h.joinRoom(client, "global")
	if client.TeamID != "" {
		h.joinRoom(client, "team:"+client.TeamID)
	}

	h.sendPresenceUpdate(client, true)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)

		for room := range client.Rooms {
			h.leaveRoom(client, room)
		}

		h.logger.Infof("Client unregistered: %s (User: %s)", client.ID, client.UserID)
		h.sendPresenceUpdate(client, false)
	}
}

func (h *Hub) joinRoom(client *Client, room string) {
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.Rooms[room] = true
	h.logger.Debugf("Client %s joined room %s", client.ID, room)
}

func (h *Hub) leaveRoom(client *Client, room string) {
	if h.rooms[room] != nil {
		delete(h.rooms[room], client)
		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(client.Rooms, room)
	h.logger.Debugf("Client %s left room %s", client.ID, room)
}

func (h *Hub) broadcastMessage(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if message.Room != "" {
		if clients, ok := h.rooms[message.Room]; ok {
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					h.logger.Warnf("Client %s send channel is full, dropping message", client.ID)
				}
			}
		}
	} else {
		for _, client := range h.clients {
			select {
			case client.Send <- data:
			default:
				h.logger.Warnf("Client %s send channel is full, dropping message", client.ID)
			}
		}
	}
}

func (h *Hub) SendToUser(userID string, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				h.logger.Warnf("Client %s send channel is full, dropping message", client.ID)
			}
		}
	}
}

func (h *Hub) SendToTeam(teamID string, message *Message) {
	message.Room = "team:" + teamID
	h.broadcast <- message
}

func (h *Hub) sendPresenceUpdate(client *Client, online bool) {
	status := "offline"
	if online {
		status = "online"
	}

	message := &Message{
		Type:      string(MessageTypePresence),
		UserID:    client.UserID,
		Data:      map[string]interface{}{"status": status},
		Timestamp: time.Now(),
	}

	if client.TeamID != "" {
		message.Room = "team:" + client.TeamID
	}

	h.broadcast <- message
}

func (h *Hub) GetOnlineUsers(teamID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userMap := make(map[string]bool)
	roomName := "team:" + teamID

	if clients, ok := h.rooms[roomName]; ok {
		for client := range clients {
			userMap[client.UserID] = true
		}
	}

	users := make([]string, 0, len(userMap))
	for userID := range userMap {
		users = append(users, userID)
	}

	return users
}