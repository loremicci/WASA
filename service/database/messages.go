package database

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"time"
)

func (db *appdbimpl) SendMessage(msg Message, replyTo string) (Message, error) {
	u, _ := uuid.NewV4()
	msg.ID = u.String()
	msg.Timestamp = time.Now()

	// Update last_read for the sender so they are marked as having read up to this point
	db.c.Exec("UPDATE conversation_members SET last_read = ?, last_received = ? WHERE conversation_id = ? AND user_id = ?", msg.Timestamp, msg.Timestamp, msg.ConversationID, msg.Sender.ID)

	// Send message
	var replyToVal interface{} = nil
	if replyTo != "" {
		replyToVal = replyTo
	}

	_, err := db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, type, timestamp, reply_to) VALUES (?, ?, ?, ?, ?, ?, ?)",
		msg.ID, msg.ConversationID, msg.Sender.ID, msg.Content, msg.Type, msg.Timestamp, replyToVal)
	if err != nil {
		return Message{}, err
	}

	msg.Status = "sent"
	return msg, nil
}

func (db *appdbimpl) DeleteMessage(messageId string, userId string) error {
	// Only delete if sender_id = userId
	_, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND sender_id = ?", messageId, userId)
	return err
}

func (db *appdbimpl) ForwardMessage(messageId string, targetConversationId string, senderId string) (Message, error) {
	// Get original message
	var content, msgType string
	err := db.c.QueryRow("SELECT content, type FROM messages WHERE id = ?", messageId).Scan(&content, &msgType)
	if err != nil {
		return Message{}, err
	}

	sender, _ := db.GetUserByID(senderId)

	msg := Message{
		ConversationID: targetConversationId,
		Sender:         sender,
		Content:        content,
		Type:           msgType,
	}

	return db.SendMessage(msg, "")
}

func (db *appdbimpl) CommentMessage(messageId string, userId string, emoticon string) error {
	_, err := db.c.Exec("INSERT OR IGNORE INTO reactions (message_id, user_id, emoticon) VALUES (?, ?, ?)", messageId, userId, emoticon)
	return err
}

func (db *appdbimpl) UncommentMessage(messageId string, userId string, emoticon string) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ? AND emoticon = ?", messageId, userId, emoticon)
	return err
}

// Helpers
func (db *appdbimpl) getReactionsForMessage(messageId string) []Reaction {
	rows, err := db.c.Query("SELECT user_id, emoticon FROM reactions WHERE message_id = ?", messageId)
	if err != nil {
		return []Reaction{}
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		var userId string
		if err := rows.Scan(&userId, &r.Emoticon); err == nil {
			u, _ := db.GetUserByID(userId)
			r.User = u
			reactions = append(reactions, r)
		}
	}
	if reactions == nil {
		return []Reaction{}
	}
	return reactions
}

func (db *appdbimpl) calculateMessageStatus(messageId string, conversationId string, timestamp time.Time) string {
	query := `
		SELECT 
			MIN(last_read >= ?) as all_read,
			MIN(last_received >= ?) as all_received
		FROM conversation_members 
		WHERE conversation_id = ? AND user_id != (SELECT sender_id FROM messages WHERE id = ?)
	`
	var allRead, allReceived sql.NullBool
	err := db.c.QueryRow(query, timestamp, timestamp, conversationId, messageId).Scan(&allRead, &allReceived)
	if err != nil {
		return "sent"
	}
	if allRead.Valid && allRead.Bool {
		return "read"
	}
	if allReceived.Valid && allReceived.Bool {
		return "received"
	}
	return "sent"
}
