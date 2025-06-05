package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	host   = "db"
	port   = 5432
	user   = "postgres"
	dbName = "bot_data"
)

var dbconn *pgxpool.Pool

// Tries to initialize the dbconn variable with a connection from connectToDatabse().
// Returns a connection to the db which should be closed when the application exits.
func Init() (*pgxpool.Pool, error) {
	db, err := connectToDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	dbconn = db
	return db, nil
}

func connectToDatabase() (*pgxpool.Pool, error) {
	// Connect to database
	password := os.Getenv("POSTGRES_PASSWORD")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbName)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	// Make sure database is responding
	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database did not respond after connecting: %w", err)
	}
	return dbpool, nil
}

func AddCodeforcesUser(discID, cfName string) error {
	tx, err := dbconn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO user_data (discord_id, codeforces_name) VALUES (%s, '%s');",
			discID, cfName))
	if err != nil {
		return fmt.Errorf("failed to insert discord id %s and codeforces name '%s' into user_data: %w",
			discID, cfName, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit insertion to user_data: %w", err)
	}
	return nil
}
