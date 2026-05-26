package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	Ping() error
	
	// User operations
	DoLogin(username string) (string, error)
	SearchUsers(query string) ([]User, error)
	GetUserByID(userId string) (User, error)
	SetMyUserName(userId string, newName string) error
	SetMyPhoto(userId string, photoPath string) error
	
	// Group operations
	CreateGroup(name string, memberIds []string) (Group, error)
	SetGroupName(groupId string, name string) error
	SetGroupPhoto(groupId string, photoPath string) error
	AddToGroup(groupId string, userId string) error
	LeaveGroup(groupId string, userId string) error
	
	// Conversation operations
	GetMyConversations(userId string) ([]Conversation, error)
	CreateConversation(userId1 string, userId2 string) (Conversation, error)
	GetConversationMessages(userId string, conversationId string) ([]Message, error)
	
	// Message operations
	SendMessage(msg Message, replyTo string) (Message, error)
	DeleteMessage(messageId string, userId string) error
	ForwardMessage(messageId string, targetConversationId string, senderId string) (Message, error)
	CommentMessage(messageId string, userId string, emoticon string) error
	UncommentMessage(messageId string, userId string, emoticon string) error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create tables if they do not exist
	err := createSchema(db)
	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
