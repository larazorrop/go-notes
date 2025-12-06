package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- MODELOS  ---
type User struct {
	ID           int
	Username     string
	PasswordHash string
}

type Note struct {
	ID        int
	UserID    int
	Content   string
	CreatedAt time.Time
}

// --- BASE DE DADOS ---
var db *pgxpool.Pool

func InitDB() error {
	// Leitura das variáveis de ambiente
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Validação simples
	if host == "" || user == "" || dbname == "" {
		return fmt.Errorf("variáveis de ambiente da DB incompletas")
	}

	// Construção da Connection String no formato Key-Value do Postgres
	// Exemplo final: "host=localhost port=5432 user=postgres password=123 dbname=gonotes sslmode=disable"
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	// Apgxpool aceita tanto URL (postgres://) como Key-Value (host=...)
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("não foi possível conectar à DB: %w", err)
	}

	// Testa a conexão efetivamente (o New apenas configura o pool)
	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("erro ao fazer ping à DB: %w", err)
	}

	// Criação automática das tabelas
	return createTables()
}

func createTables() error {
	schema := `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS notes (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            content TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`
	_, err := db.Exec(context.Background(), schema)
	return err
}

// --- FUNÇÕES DE ACESSO A DADOS ---

func CreateUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), "INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, string(hash))
	return err
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow(context.Background(), "SELECT id, username, password_hash FROM users WHERE username=$1", username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	return &user, err
}

func CreateNote(userID int, content string) error {
	_, err := db.Exec(context.Background(), "INSERT INTO notes (user_id, content) VALUES ($1, $2)", userID, content)
	return err
}

func GetNotesByUserID(userID int) ([]Note, error) {
	rows, err := db.Query(context.Background(), "SELECT id, content, created_at FROM notes WHERE user_id=$1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func DeleteNote(noteID, userID int) error {
	_, err := db.Exec(context.Background(), "DELETE FROM notes WHERE id=$1 AND user_id=$2", noteID, userID)
	return err
}
