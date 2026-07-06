package database

import (
	"fmt"
	"path/filepath"

	sqlite "gosqlite.org"
	"gosqlite.org/vfs/crypto"
)

type Database struct {
	db *sqlite.DB
}

func NewDatabase(encryptionKey []byte, storagePath string) (*Database, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("invalid encryption key")
	}

	dbPath := filepath.Join(storagePath, "bloom.db")

	db, err := crypto.Open(
		sqlite.Config{
			Path:    dbPath,
			Pragmas: sqlite.RecommendedPragmas(),
		},
		crypto.Options{
			Key: encryptionKey,
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

	CREATE TABLE IF NOT EXISTS users (
	    id TEXT PRIMARY KEY,
	    username TEXT NOT NULL UNIQUE,
	    display_name TEXT,
	    description TEXT,
	    ml_kem_public_key TEXT NOT NULL,
	    ecdh_public_key TEXT NOT NULL,
	    ed_public_key TEXT NOT NULL,
	    date DATETIME NOT NULL
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
