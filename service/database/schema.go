package database

import "database/sql"

func createSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			photo_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT,
			photo_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS conversation_members (
			conversation_id TEXT,
			user_id TEXT,
			last_received DATETIME,
			last_read DATETIME,
			PRIMARY KEY (conversation_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			content TEXT,
			photo TEXT,
			type TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			reply_to TEXT,
			is_forwarded BOOLEAN DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS reactions (
			message_id TEXT,
			user_id TEXT,
			emoticon TEXT,
			PRIMARY KEY (message_id, user_id, emoticon)
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Add is_online column safely for existing DBs
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN is_online BOOLEAN DEFAULT 0")

	return nil
}
