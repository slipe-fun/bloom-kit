package mobile

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/slipe-fun/bloom-kit/domain"
	sqlite "gosqlite.org"
	"gosqlite.org/vfs/crypto"
)

type Database struct {
	db *sqlite.DB
}

func (c *BloomClient) newDatabase() (*Database, error) {
	if len(c.encryptionKey) != 32 {
		return nil, fmt.Errorf("invalid encryption key")
	}

	dbPath := filepath.Join(c.storagePath, "bloom.db")

	db, err := crypto.Open(
		sqlite.Config{
			Path:    dbPath,
			Pragmas: sqlite.RecommendedPragmas(),
		},
		crypto.Options{
			Key: c.encryptionKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open encrypted database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS chats (
	    id INTEGER PRIMARY KEY,
	    members TEXT NOT NULL,
	    handshake TEXT NOT NULL,
	    chat_key TEXT NOT NULL,
	    sync_key TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) closeDatabase() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

func (d *Database) saveChat(chat domain.Chat, chatKey, syncKey []byte) error {
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

func (d *Database) getChat(chatID int) (*domain.ChatWithKeys, error) {
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

func (d *Database) getChats() ([]domain.ChatWithKeys, error) {
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
