package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (d *Database) SaveMessage(message *domain.Message) error {
	if message.ReplyToMessage != nil {
		message.ReplyToMessage.ChatID = message.ChatID

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
			ON CONFLICT(id) DO UPDATE SET
				chat_id = excluded.chat_id,
				author_id = excluded.author_id,
				seen = excluded.seen,
				reply_to = excluded.reply_to,
				nonce = excluded.nonce,
				content = excluded.content;
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
		ON CONFLICT(id) DO UPDATE SET
			chat_id = excluded.chat_id,
			author_id = excluded.author_id,
			seen = excluded.seen,
			reply_to = excluded.reply_to,
			nonce = excluded.nonce,
			content = excluded.content;
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

func (d *Database) SaveMessages(messages []domain.Message) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := range messages {
		msg := &messages[i]

		if msg.ReplyToMessage != nil {
			msg.ReplyToMessage.ChatID = msg.ChatID

			_, err := tx.Exec(`
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
				ON CONFLICT(id) DO UPDATE SET
					chat_id = excluded.chat_id,
					author_id = excluded.author_id,
					seen = excluded.seen,
					reply_to = excluded.reply_to,
					nonce = excluded.nonce,
					content = excluded.content;
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
			ON CONFLICT(id) DO UPDATE SET
				chat_id = excluded.chat_id,
				author_id = excluded.author_id,
				seen = excluded.seen,
				reply_to = excluded.reply_to,
				nonce = excluded.nonce,
				content = excluded.content;
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

func (d *Database) GetMessage(messageID int) (*domain.Message, error) {
	row := d.db.QueryRow(`
		SELECT id, chat_id, author_id, seen, reply_to, nonce, content
		FROM messages
		WHERE id = ?
	`, messageID)

	var msg domain.Message
	err := row.Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.AuthorID,
		&msg.Seen,
		&msg.ReplyTo,
		&msg.Nonce,
		&msg.Content,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	slice := []domain.Message{msg}
	if err := d.populateReplies(slice); err != nil {
		return nil, err
	}

	return &slice[0], nil
}

func (d *Database) GetMessages(chatID, beforeID, limit int) ([]domain.Message, error) {
	var query string
	var args []any

	if beforeID > 0 {
		query = `
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
			ORDER BY id ASC
			LIMIT ?
		`
		args = []any{chatID, beforeID, limit}
	} else {
		query = `
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
			ORDER BY id ASC
			LIMIT ?
		`
		args = []any{chatID, limit}
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
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
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := d.populateReplies(messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *Database) GetChatLastMessage(chatID int) (*domain.Message, error) {
	row := d.db.QueryRow(`
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
		ORDER BY id DESC
		LIMIT 1
	`, chatID)

	var msg domain.Message
	err := row.Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.AuthorID,
		&msg.Seen,
		&msg.ReplyTo,
		&msg.Nonce,
		&msg.Content,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	slice := []domain.Message{msg}
	if err := d.populateReplies(slice); err != nil {
		return nil, err
	}

	return &slice[0], nil
}

func (d *Database) GetChatsLastMessages(chatIDs []int) (map[int]*domain.Message, error) {
	result := make(map[int]*domain.Message)
	if len(chatIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(chatIDs))
	args := make([]any, len(chatIDs))
	for i, id := range chatIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		WITH ranked_messages AS (
			SELECT
				id,
				chat_id,
				author_id,
				seen,
				reply_to,
				nonce,
				content,
				ROW_NUMBER() OVER (PARTITION BY chat_id ORDER BY id DESC) as rn
			FROM messages
			WHERE chat_id IN (%s)
		)
		SELECT id, chat_id, author_id, seen, reply_to, nonce, content
		FROM ranked_messages
		WHERE rn = 1
	`, strings.Join(placeholders, ","))

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
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
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := d.populateReplies(messages); err != nil {
		return nil, err
	}

	for i := range messages {
		result[messages[i].ChatID] = &messages[i]
	}

	return result, nil
}

func (d *Database) populateReplies(messages []domain.Message) error {
	replyIDsMap := make(map[int]bool)
	for _, msg := range messages {
		if msg.ReplyTo != nil {
			replyIDsMap[*msg.ReplyTo] = true
		}
	}

	if len(replyIDsMap) == 0 {
		return nil
	}

	replyMap := make(map[int]*domain.MessageWithDecryptedData, len(replyIDsMap))
	ids := make([]any, 0, len(replyIDsMap))
	placeholders := make([]string, 0, len(replyIDsMap))

	for id := range replyIDsMap {
		ids = append(ids, id)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(`
		SELECT id, chat_id, author_id, seen, reply_to, nonce, content
		FROM messages
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	replyRows, err := d.db.Query(query, ids...)
	if err != nil {
		return err
	}
	defer replyRows.Close()

	for replyRows.Next() {
		reply := new(domain.MessageWithDecryptedData)
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
			return err
		}

		replyMap[reply.ID] = reply
	}

	if err := replyRows.Err(); err != nil {
		return err
	}

	for i := range messages {
		if messages[i].ReplyTo != nil {
			if replyMsg, exists := replyMap[*messages[i].ReplyTo]; exists {
				messages[i].ReplyToMessage = replyMsg
			}
		}
	}

	return nil
}
