package database

import (
	"github.com/gofrs/uuid"
	"time"
)

func (db *appdbimpl) CreateGroup(name string, memberIds []string) (Group, error) {
	u, _ := uuid.NewV4()
	id := u.String()

	tx, err := db.c.Begin()
	if err != nil {
		return Group{}, err
	}
	defer tx.Rollback()

	// Insert into conversations
	_, err = tx.Exec("INSERT INTO conversations (id, type, name, photo_url) VALUES (?, 'group', ?, '')", id, name)
	if err != nil {
		return Group{}, err
	}

	now := time.Now()

	for _, memberId := range memberIds {
		_, err = tx.Exec("INSERT INTO conversation_members (conversation_id, user_id, last_received, last_read) VALUES (?, ?, ?, ?)", id, memberId, now, now)
		if err != nil {
			return Group{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Group{}, err
	}

	// Fetch members
	members := []User{}
	for _, memberId := range memberIds {
		u, _ := db.GetUserByID(memberId)
		members = append(members, u)
	}

	return Group{
		ID:       id,
		Name:     name,
		PhotoURL: "",
		Members:  members,
	}, nil
}

func (db *appdbimpl) SetGroupName(groupId string, name string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? AND type = 'group'", name, groupId)
	return err
}

func (db *appdbimpl) SetGroupPhoto(groupId string, photoPath string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ? AND type = 'group'", photoPath, groupId)
	return err
}

func (db *appdbimpl) AddToGroup(groupId string, userId string) error {
	now := time.Now()
	_, err := db.c.Exec("INSERT OR IGNORE INTO conversation_members (conversation_id, user_id, last_received, last_read) VALUES (?, ?, ?, ?)", groupId, userId, now, now)
	return err
}

func (db *appdbimpl) LeaveGroup(groupId string, userId string) error {
	_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", groupId, userId)
	return err
}
