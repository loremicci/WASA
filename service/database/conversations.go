package database

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"sort"
	"time"
)

func (db *appdbimpl) GetMyConversations(userId string) ([]Conversation, error) {
	query := `
		SELECT c.id, c.type, c.name, c.photo_url,
			(SELECT timestamp FROM messages m WHERE m.conversation_id = c.id ORDER BY timestamp DESC LIMIT 1),
			(SELECT content FROM messages m WHERE m.conversation_id = c.id ORDER BY timestamp DESC LIMIT 1),
			(SELECT type FROM messages m WHERE m.conversation_id = c.id ORDER BY timestamp DESC LIMIT 1)
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = ?
	`
	rows, err := db.c.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var c Conversation
		var cType string
		var cName sql.NullString
		var cPhoto sql.NullString
		var ts sql.NullTime
		var preview sql.NullString
		var msgType sql.NullString

		if err := rows.Scan(&c.ID, &cType, &cName, &cPhoto, &ts, &preview, &msgType); err != nil {
			return nil, err
		}

		c.IsGroup = cType == "group"
		if c.IsGroup {
			c.Name = cName.String
			c.PhotoURL = cPhoto.String
		} else {
			// Find the other user
			otherUserQuery := `
				SELECT u.name, u.photo_url 
				FROM users u 
				JOIN conversation_members cm ON u.id = cm.user_id 
				WHERE cm.conversation_id = ? AND cm.user_id != ? LIMIT 1
			`
			var otherName, otherPhoto sql.NullString
			err := db.c.QueryRow(otherUserQuery, c.ID, userId).Scan(&otherName, &otherPhoto)
			if err == nil {
				c.Name = otherName.String
				c.PhotoURL = otherPhoto.String
			}
		}

		if ts.Valid {
			c.LatestMessageTimestamp = ts.Time
		}
		if preview.Valid {
			if msgType.String == "photo" {
				c.LatestMessagePreview = "[Photo]"
			} else {
				// Trim long text preview
				prev := preview.String
				if len(prev) > 50 {
					prev = prev[:47] + "..."
				}
				c.LatestMessagePreview = prev
			}
		}
		convs = append(convs, c)
	}

	// Sort by LatestMessageTimestamp descending
	sort.Slice(convs, func(i, j int) bool {
		return convs[i].LatestMessageTimestamp.After(convs[j].LatestMessageTimestamp)
	})

	return convs, nil
}

func (db *appdbimpl) CreateConversation(userId1 string, userId2 string) (Conversation, error) {
	// Check if already exists
	query := `
		SELECT c.id 
		FROM conversations c
		JOIN conversation_members cm1 ON c.id = cm1.conversation_id
		JOIN conversation_members cm2 ON c.id = cm2.conversation_id
		WHERE c.type = 'direct' AND cm1.user_id = ? AND cm2.user_id = ?
	`
	var existingId string
	err := db.c.QueryRow(query, userId1, userId2).Scan(&existingId)
	if err == nil {
		// Existing conversation, return it by fetching full object
		// We can just construct a basic Conversation object to return, or run a full query
		var c Conversation
		c.ID = existingId
		c.IsGroup = false
		// get the other user for Name
		otherUser, _ := db.GetUserByID(userId2)
		c.Name = otherUser.Name
		c.PhotoURL = otherUser.PhotoURL
		return c, nil
	}
	if err != sql.ErrNoRows {
		return Conversation{}, err
	}

	u, _ := uuid.NewV4()
	id := u.String()

	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO conversations (id, type, name, photo_url) VALUES (?, 'direct', '', '')", id)
	if err != nil {
		return Conversation{}, err
	}

	now := time.Now()
	_, err = tx.Exec("INSERT INTO conversation_members (conversation_id, user_id, last_received, last_read) VALUES (?, ?, ?, ?)", id, userId1, now, now)
	if err != nil {
		return Conversation{}, err
	}
	_, err = tx.Exec("INSERT INTO conversation_members (conversation_id, user_id, last_received, last_read) VALUES (?, ?, ?, ?)", id, userId2, now, now)
	if err != nil {
		return Conversation{}, err
	}

	if err := tx.Commit(); err != nil {
		return Conversation{}, err
	}

	otherUser, _ := db.GetUserByID(userId2)
	return Conversation{
		ID:       id,
		IsGroup:  false,
		Name:     otherUser.Name,
		PhotoURL: otherUser.PhotoURL,
	}, nil
}

func (db *appdbimpl) GetConversationMessages(userId string, conversationId string) ([]Message, error) {
	// Update last_read for the user since they are opening it
	now := time.Now()
	db.c.Exec("UPDATE conversation_members SET last_read = ? WHERE conversation_id = ? AND user_id = ?", now, conversationId, userId)

	// Fetch messages
	query := `
		SELECT id, sender_id, content, type, timestamp, reply_to 
		FROM messages 
		WHERE conversation_id = ? 
		ORDER BY timestamp DESC
	`
	rows, err := db.c.Query(query, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var senderId string
		var replyTo sql.NullString
		if err := rows.Scan(&m.ID, &senderId, &m.Content, &m.Type, &m.Timestamp, &replyTo); err != nil {
			return nil, err
		}
		m.ConversationID = conversationId
		
		sender, _ := db.GetUserByID(senderId)
		m.Sender = sender

		// Need to fetch reactions
		m.Reactions = db.getReactionsForMessage(m.ID)

		// Calculate status (read/received) if user is sender
		if senderId == userId {
			m.Status = db.calculateMessageStatus(m.ID, conversationId, m.Timestamp)
		} else {
			m.Status = "received"
		}

		msgs = append(msgs, m)
	}

	return msgs, nil
}
