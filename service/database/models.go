package database

import "time"

type User struct {
	ID       string `json:"identifier"`
	Name     string `json:"username"`
	PhotoURL string `json:"photoUrl"`
}

type Group struct {
	ID       string `json:"identifier"`
	Name     string `json:"name"`
	PhotoURL string `json:"photoUrl"`
	Members  []User `json:"members"`
}

type Conversation struct {
	ID                     string    `json:"identifier"`
	IsGroup                bool      `json:"isGroup"`
	Name                   string    `json:"name"`
	PhotoURL               string    `json:"photoUrl"`
	LatestMessageTimestamp time.Time `json:"latestMessageTimestamp"`
	LatestMessagePreview   string    `json:"latestMessagePreview"`
}

type Reaction struct {
	Emoticon string `json:"emoticon"`
	User     User   `json:"user"`
}

type Message struct {
	ID             string     `json:"identifier"`
	ConversationID string     `json:"conversationId"`
	Sender         User       `json:"sender"`
	Timestamp      time.Time  `json:"timestamp"`
	Type           string     `json:"type"` // "text" or "photo"
	Content        string     `json:"content"`
	Photo          string     `json:"photo"`
	Status         string     `json:"status"` // "sent", "received", "read"
	Forwarded      bool       `json:"forwarded"`
	ReplyTo        string     `json:"replyTo,omitempty"`
	Reactions      []Reaction `json:"reactions"`
}
