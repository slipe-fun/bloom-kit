package database

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (d *Database) SaveUser(user *domain.User) error {
	_, err := d.db.Exec(`
		INSERT INTO users (id, username, display_name, description, ml_kem_public_key, ecdh_public_key, ed_public_key, date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		    username = excluded.username,
		    display_name = excluded.display_name,
		    description = excluded.description,
		    ml_kem_public_key = excluded.ml_kem_public_key,
		    ecdh_public_key = excluded.ecdh_public_key,
		    ed_public_key = excluded.ed_public_key,
		    date = excluded.date;
	`, user.ID, user.Username, user.DisplayName, user.Description, user.MlKemPublicKey, user.EcdhPublicKey, user.EdPublicKey, user.Date)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) SaveUsers(users []domain.User) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO users (id, username, display_name, description, ml_kem_public_key, ecdh_public_key, ed_public_key, date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			display_name = excluded.display_name,
			description = excluded.description,
			ml_kem_public_key = excluded.ml_kem_public_key,
			ecdh_public_key = excluded.ecdh_public_key,
			ed_public_key = excluded.ed_public_key,
			date = excluded.date;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		_, err := stmt.Exec(
			user.ID,
			user.Username,
			user.DisplayName,
			user.Description,
			user.MlKemPublicKey,
			user.EcdhPublicKey,
			user.EdPublicKey,
			user.Date,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) GetUser(userID string) (*domain.User, error) {
	row := d.db.QueryRow(`
		SELECT id, username, display_name, description, ml_kem_public_key, ecdh_public_key, ed_public_key, date
		FROM users
		WHERE id = ?
	`, userID)

	var user domain.User

	err := row.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Description, &user.MlKemPublicKey, &user.EcdhPublicKey, &user.EdPublicKey, &user.Date)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return &user, nil
}
