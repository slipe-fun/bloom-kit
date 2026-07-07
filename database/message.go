package database

import (
	"github.com/slipe-fun/bloom-kit/domain"
)

func (d *Database) SaveMessage(message *domain.MessageWithDecryptedData) error {
	_, err := d.db.Exec(`
		INSERT INTO messages (
			id,
			chat_id,
			author_id,
			seen,
			reply_to,
			nonce,
			content
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		message.ID,
		message.ChatID,
		message.AuthorID,
		message.Seen,
		message.ReplyTo,
		message.Nonce,
		message.Content,
	)

	return err
}

func (d *Database) GetMessages(chatID, afterID, limit int) ([]domain.MessageWithDecryptedData, error) {
	rows, err := d.db.Query(`
		SELECT
			id,
			chat_id,
			author_id,
			seen,
			reply_to,
			nonce,
			content
		FROM messages
		WHERE chat_id = ?
		  AND id > ?
		ORDER BY id ASC
		LIMIT ?
	`, chatID, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.MessageWithDecryptedData

	for rows.Next() {
		var msg domain.MessageWithDecryptedData

		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.AuthorID,
			&msg.Seen,
			&msg.ReplyTo,
			&msg.Nonce,
			&msg.Content,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
