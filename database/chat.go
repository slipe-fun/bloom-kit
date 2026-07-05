package database

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (d *Database) SaveChat(chat domain.Chat, chatKey, syncKey []byte) error {
	membersJSON, err := json.Marshal(chat.Members)
	if err != nil {
		return err
	}

	handshakeJSON, err := json.Marshal(chat.Handshake)
	if err != nil {
		return err
	}

	encodedChatKey := base64.StdEncoding.EncodeToString(chatKey)
	encodedSyncKey := base64.StdEncoding.EncodeToString(syncKey)

	_, err = d.db.Exec(`
		INSERT INTO chats (id, members, handshake, chat_key, sync_key)
		VALUES (?, ?, ?, ?, ?)
	`, chat.ID, string(membersJSON), string(handshakeJSON), encodedChatKey, encodedSyncKey)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetChat(chatID int) (*domain.ChatWithKeys, error) {
	row := d.db.QueryRow(`
		SELECT id, members, handshake, chat_key, sync_key
		FROM chats
		WHERE id = ?
	`, chatID)

	var (
		chat           domain.RawChat
		membersJSON    string
		handshakeJSON  string
		encodedChatKey string
		encodedSyncKey string
	)

	err := row.Scan(&chat.ID, &membersJSON, &handshakeJSON, &encodedChatKey, &encodedSyncKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(membersJSON), &chat.Members); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(handshakeJSON), &chat.Handshake); err != nil {
		return nil, err
	}

	chatKeyBytes, err := base64.StdEncoding.DecodeString(encodedChatKey)
	if err != nil {
		return nil, err
	}

	syncKeyBytes, err := base64.StdEncoding.DecodeString(encodedSyncKey)
	if err != nil {
		return nil, err
	}

	return &domain.ChatWithKeys{
		RawChat: chat,
		ChatKey: chatKeyBytes,
		SyncKey: syncKeyBytes,
	}, nil
}

func (d *Database) GetChats() ([]domain.ChatWithKeys, error) {
	rows, err := d.db.Query(`
		SELECT id, members, handshake, chat_key, sync_key
		FROM chats
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []domain.ChatWithKeys

	for rows.Next() {
		var (
			chat           domain.RawChat
			membersJSON    string
			handshakeJSON  string
			encodedChatKey string
			encodedSyncKey string
		)

		if err := rows.Scan(
			&chat.ID,
			&membersJSON,
			&handshakeJSON,
			&encodedChatKey,
			&encodedSyncKey,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(membersJSON), &chat.Members); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(handshakeJSON), &chat.Handshake); err != nil {
			return nil, err
		}

		chatKeyBytes, err := base64.StdEncoding.DecodeString(encodedChatKey)
		if err != nil {
			return nil, err
		}

		syncKeyBytes, err := base64.StdEncoding.DecodeString(encodedSyncKey)
		if err != nil {
			return nil, err
		}

		chats = append(chats, domain.ChatWithKeys{
			RawChat: chat,
			ChatKey: chatKeyBytes,
			SyncKey: syncKeyBytes,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (d *Database) EditChatMembers(chatID int, newMembers []User) error {
	membersJSON, err := json.Marshal(newMembers)
	if err != nil {
		return err
	}

	query := `UPDATE chats SET members = ? WHERE id = ?;`
	result, err := d.db.Exec(query, string(membersJSON), chatID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("chat with ID %d not found", chatID)
	}

	return nil
}
