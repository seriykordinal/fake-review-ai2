package database

import (
	"database/sql"
	"fake-review-ai2/models"
	"time"
)

func CreateUserTable() error {
	// Таблица должна уже существовать, но для полноты
	_, err := DB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            role VARCHAR(20) DEFAULT 'user',
            is_verified BOOLEAN DEFAULT FALSE,
            verification_code VARCHAR(10),
            code_expires_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT NOW()
        )`)
	return err
}

func InsertUser(email, passwordHash, code string, expiresAt time.Time) error {
	_, err := DB.Exec(`
        INSERT INTO users (email, password_hash, verification_code, code_expires_at)
        VALUES ($1, $2, $3, $4)`,
		email, passwordHash, code, expiresAt)
	return err
}

func FindUserByEmail(email string) (*models.User, error) {
	var u models.User
	var verificationCode sql.NullString
	var codeExpiresAt sql.NullTime
	row := DB.QueryRow(`
        SELECT id, email, password_hash, role, is_verified, verification_code, code_expires_at, created_at
        FROM users WHERE email = $1`, email)
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified, &verificationCode, &codeExpiresAt, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if verificationCode.Valid {
		u.VerificationCode = verificationCode.String
	}
	if codeExpiresAt.Valid {
		u.CodeExpiresAt = codeExpiresAt.Time
	}
	return &u, nil
}

func UpdateVerificationCode(email, code string, expiresAt time.Time) error {
	_, err := DB.Exec(`
        UPDATE users SET verification_code = $1, code_expires_at = $2
        WHERE email = $3`, code, expiresAt, email)
	return err
}

func VerifyUser(email string) error {
	_, err := DB.Exec(`
        UPDATE users SET is_verified = TRUE, verification_code = NULL, code_expires_at = NULL
        WHERE email = $1`, email)
	return err
}

// === Админские функции ===
func GetAllUsers() ([]models.User, error) {
	rows, err := DB.Query(`
        SELECT id, email, role, is_verified, created_at
        FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsVerified, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func DeleteUser(userID int) error {
	_, err := DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	return err
}

func UpdateUserRole(userID int, role string) error {
	_, err := DB.Exec(`UPDATE users SET role = $1 WHERE id = $2`, role, userID)
	return err
}

func FindUserByID(userID int) (*models.User, error) {
	var u models.User
	row := DB.QueryRow(`
        SELECT id, email, role, is_verified, created_at
        FROM users WHERE id = $1`, userID)
	err := row.Scan(&u.ID, &u.Email, &u.Role, &u.IsVerified, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func DeleteUnverifiedUsers() error {
	_, err := DB.Exec(`DELETE FROM users WHERE is_verified = false`)
	return err
}

func InsertUserVerified(email, passwordHash string) error {
	_, err := DB.Exec(`
        INSERT INTO users (email, password_hash, is_verified)
        VALUES ($1, $2, TRUE)`,
		email, passwordHash)
	return err
}
