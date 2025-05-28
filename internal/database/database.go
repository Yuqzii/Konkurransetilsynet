package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

const (
	host   = "db"
	port   = 5432
	user   = "postgres"
	dbName = "bot-data"
)

var dbConn *pgx.Conn

// Tries to initiliazes this packages connection variable.
// Returns a connection to the db which should be closed when the application exits.
func InitDB() (*pgx.Conn, error) {
	conn, err := connectToDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	dbConn = conn
	return conn, nil
}

func connectToDatabase() (*pgx.Conn, error) {
	// Connect to database
	password := os.Getenv("POSTGRES_PASSWORD")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbName)
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	// Make sure database is responding
	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database did not respond after connecting: %w", err)
	}
	return db, nil
}

func AddUser(discID, cfName string, conn *pgx.Conn) error {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO %s (discordID, username) VALUES (%s, %s);", dbName, discID, cfName))
	if err != nil {
		return fmt.Errorf("failed to insert discord id %s and username %s into %s: %w", discID, cfName, dbName, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit insertion to %s: %w", dbName, err)
	}
	return nil
}
