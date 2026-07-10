package database

import (
	"fmt"
	"strings"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (d *Database) SaveMessage(message *domain.Message) error {
	if message.ReplyToMessage != nil {
		_, err := d.db.Exec(`
			INSERT OR REPLACE INTO messages (
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
			message.ReplyToMessage.ID,
			message.ReplyToMessage.ChatID,
			message.ReplyToMessage.AuthorID,
			message.ReplyToMessage.Seen,
			message.ReplyToMessage.ReplyTo,
			message.ReplyToMessage.Nonce,
			message.ReplyToMessage.Content,
		)
		if err != nil {
			return err
		}
	}

	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO messages (
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

func (d *Database) GetMessages(chatID, beforeID, limit int) ([]domain.Message, error) {
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
		  AND id < ?
		ORDER BY id DESC
		LIMIT ?
	`, chatID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	replyIDsMap := make(map[int]bool)

	for rows.Next() {
		var msg domain.Message

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

		if msg.ReplyTo != nil {
			replyIDsMap[*msg.ReplyTo] = true
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(replyIDsMap) > 0 {
		replyMap := make(map[int]*domain.MessageWithDecryptedData, len(replyIDsMap))

		ids := make([]any, 0, len(replyIDsMap))
		placeholders := make([]string, 0, len(replyIDsMap))

		for id := range replyIDsMap {
			ids = append(ids, id)
			placeholders = append(placeholders, "?")
		}

		query := fmt.Sprintf(`
			SELECT
				id,
				chat_id,
				author_id,
				seen,
				reply_to,
				nonce,
				content
			FROM messages
			WHERE id IN (%s)
		`, strings.Join(placeholders, ","))

		replyRows, err := d.db.Query(query, ids...)
		if err != nil {
			return nil, err
		}
		defer replyRows.Close()

		for replyRows.Next() {
			var reply domain.MessageWithDecryptedData
			err := replyRows.Scan(
				&reply.ID,
				&reply.ChatID,
				&reply.AuthorID,
				&reply.Seen,
				&reply.ReplyTo,
				&reply.Nonce,
				&reply.Content,
			)
			if err != nil {
				return nil, err
			}

			replyCopy := reply
			replyMap[reply.ID] = &replyCopy
		}

		if err := replyRows.Err(); err != nil {
			return nil, err
		}

		for i := range messages {
			if messages[i].ReplyTo != nil {
				if replyMsg, exists := replyMap[*messages[i].ReplyTo]; exists {
					messages[i].ReplyToMessage = replyMsg
				}
			}
		}
	}

	return messages, nil
}

func (d *Database) SaveMessages(messages []domain.Message) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := range messages {
		msg := &messages[i]

		if msg.ReplyToMessage != nil {
			_, err := tx.Exec(`
				INSERT OR REPLACE INTO messages (
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
				msg.ReplyToMessage.ID,
				msg.ReplyToMessage.ChatID,
				msg.ReplyToMessage.AuthorID,
				msg.ReplyToMessage.Seen,
				msg.ReplyToMessage.ReplyTo,
				msg.ReplyToMessage.Nonce,
				msg.ReplyToMessage.Content,
			)
			if err != nil {
				return err
			}
		}

		_, err := tx.Exec(`
			INSERT OR REPLACE INTO messages (
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
			msg.ID,
			msg.ChatID,
			msg.AuthorID,
			msg.Seen,
			msg.ReplyTo,
			msg.Nonce,
			msg.Content,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
