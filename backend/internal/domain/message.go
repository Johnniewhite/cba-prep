package domain

import (
	"time"
)

type Message struct {
	ID          string       `json:"id" db:"id"`
	TeamID      string       `json:"team_id" db:"team_id"`
	ChannelID   string       `json:"channel_id" db:"channel_id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Content     string       `json:"content" db:"content"`
	Type        MessageType  `json:"type" db:"type"`
	Attachments []Attachment `json:"attachments,omitempty"`
	IsEdited    bool         `json:"is_edited" db:"is_edited"`
	IsDeleted   bool         `json:"is_deleted" db:"is_deleted"`
	ReplyToID   *string      `json:"reply_to_id,omitempty" db:"reply_to_id"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type Attachment struct {
	ID        string `json:"id" db:"id"`
	MessageID string `json:"message_id" db:"message_id"`
	FileName  string `json:"file_name" db:"file_name"`
	FileSize  int64  `json:"file_size" db:"file_size"`
	FileType  string `json:"file_type" db:"file_type"`
	URL       string `json:"url" db:"url"`
}

type Channel struct {
	ID          string      `json:"id" db:"id"`
	TeamID      string      `json:"team_id" db:"team_id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        ChannelType `json:"type" db:"type"`
	IsPrivate   bool        `json:"is_private" db:"is_private"`
	CreatedBy   string      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type ChannelType string

const (
	ChannelTypeGeneral ChannelType = "general"
	ChannelTypeRandom  ChannelType = "random"
	ChannelTypeCustom  ChannelType = "custom"
	ChannelTypeDirect  ChannelType = "direct"
)

type CreateMessage struct {
	ChannelID string      `json:"channel_id" validate:"required"`
	Content   string      `json:"content" validate:"required,min=1,max=4000"`
	Type      MessageType `json:"type" validate:"required,oneof=text image file"`
	ReplyToID *string     `json:"reply_to_id,omitempty"`
}

type UpdateMessage struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

type CreateChannel struct {
	TeamID      string      `json:"team_id" validate:"required"`
	Name        string      `json:"name" validate:"required,min=1,max=100"`
	Description string      `json:"description" validate:"max=500"`
	Type        ChannelType `json:"type" validate:"required,oneof=general random custom"`
	IsPrivate   bool        `json:"is_private"`
}