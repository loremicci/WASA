package database

import (
	"database/sql"
	"errors"
	"github.com/gofrs/uuid"
)

func (db *appdbimpl) DoLogin(username string) (string, error) {
	var id string
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", username).Scan(&id)
	if err == nil {
		_, _ = db.c.Exec("UPDATE users SET is_online = 1 WHERE id = ?", id)
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	// Create user
	u, _ := uuid.NewV4()
	id = u.String()
	_, err = db.c.Exec("INSERT INTO users (id, name, photo_url, is_online) VALUES (?, ?, '', 1)", id, username)
	return id, err
}

func (db *appdbimpl) DoLogout(userId string) error {
	_, err := db.c.Exec("UPDATE users SET is_online = 0 WHERE id = ?", userId)
	return err
}

func (db *appdbimpl) SearchUsers(query string) ([]User, error) {
	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = db.c.Query("SELECT id, name, photo_url, is_online FROM users")
	} else {
		rows, err = db.c.Query("SELECT id, name, photo_url, is_online FROM users WHERE name LIKE '%' || ? || '%'", query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.PhotoURL, &u.IsOnline); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (db *appdbimpl) SetMyUserName(userId string, newName string) error {
	var id string
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", newName).Scan(&id)
	if err == nil && id != userId {
		return errors.New("username already taken")
	}
	_, err = db.c.Exec("UPDATE users SET name = ? WHERE id = ?", newName, userId)
	return err
}

func (db *appdbimpl) SetMyPhoto(userId string, photoPath string) error {
	_, err := db.c.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoPath, userId)
	return err
}

func (db *appdbimpl) GetUserByID(id string) (User, error) {
	var u User
	err := db.c.QueryRow("SELECT id, name, photo_url, is_online FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.PhotoURL, &u.IsOnline)
	return u, err
}
