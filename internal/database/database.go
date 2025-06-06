package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
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

func DiscordIDExists(discID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT discord_id FROM user_data WHERE discord_id='%s';", discID)
	var dbDiscID string
	err := dbconn.QueryRow(context.Background(), queryStr).Scan(&dbDiscID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func AddCodeforcesUser(discID, handle string) error {
	tx, err := dbconn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	queryStr := fmt.Sprintf("INSERT INTO user_data (discord_id, codeforces_name) VALUES (%s, '%s');",
		discID, handle)
	_, err = tx.Exec(context.Background(), queryStr)
	if err != nil {
		return fmt.Errorf("failed to insert discord id %s and codeforces name '%s' into user_data: %w",
			discID, handle, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit insertion '%s' to user_data: %w", queryStr, err)
	}
	return nil
}

func UpdateCodeforcesUser(discID, handle string) error {
	tx, err := dbconn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	queryStr := fmt.Sprintf("UPDATE user_data SET codeforces_name='%s' WHERE discord_id=%s;", handle, discID)
	_, err = tx.Exec(context.Background(), queryStr)
	if err != nil {
		return fmt.Errorf("failed to update the codeforces handle belonging to discord id %s to '%s': %w",
			discID, handle, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit update '%s' to user_data: %w", queryStr, err)
	}
	return nil
}

func GetConnectedCodeforces(discID, handle string) (connectedHandle string, err error) {
	queryStr := fmt.Sprintf("SELECT codeforces_name FROM user_data WHERE discord_id='%s';", discID)
	err = dbconn.QueryRow(context.Background(), queryStr).Scan(&connectedHandle)
	if err != nil {
		return "", err
	}

	return connectedHandle, nil
}
