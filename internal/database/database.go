package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuqzii/konkurransetilsynet/internal/codeforces"
)

type db struct {
	conn *pgxpool.Pool
}

// Creates a new db with the provided parameters.
// Remember to close using db.Close().
func New(ctx context.Context, host, user, password, dbName string, port uint16) (*db, error) {
	conn, err := connectToDatabase(ctx, host, user, password, dbName, port)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &db{conn: conn}, nil
}

// Should be called when application exits.
func (db *db) Close() {
	db.conn.Close()
}

func connectToDatabase(ctx context.Context, host, user, password, dbName string, port uint16) (*pgxpool.Pool, error) {
	// Connect to database.
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbName)
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	// Make sure database is responding.
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("database did not respond after connecting: %w", err)
	}
	return dbpool, nil
}

func (db *db) DiscordIDExists(ctx context.Context, discID string) (bool, error) {
	var dbDiscID string
	err := db.conn.QueryRow(ctx,
		"SELECT discord_id FROM user_data WHERE discord_id=$1;", discID).Scan(&dbDiscID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (db *db) AddCodeforcesUser(ctx context.Context, discID, handle string) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) // nolint: errcheck

	_, err = tx.Exec(ctx,
		"INSERT INTO user_data (discord_id, codeforces_handle) VALUES ($1, $2);", discID, handle)
	if err != nil {
		return fmt.Errorf("failed to insert discord id %s and codeforces name '%s' into user_data: %w",
			discID, handle, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit insertion to user_data: %w", err)
	}
	return nil
}

func (db *db) UpdateCodeforcesUser(ctx context.Context, discID, handle string) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) // nolint: errcheck

	_, err = tx.Exec(ctx,
		"UPDATE user_data SET codeforces_handle=$1 WHERE discord_id=$2;", handle, discID)
	if err != nil {
		return fmt.Errorf("failed to update the codeforces handle belonging to discord id %s to '%s': %w",
			discID, handle, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit update to user_data: %w", err)
	}
	return nil
}

func (db *db) GetConnectedCodeforces(ctx context.Context, discID string) (connectedHandle string, err error) {
	err = db.conn.QueryRow(ctx,
		"SELECT codeforces_handle FROM user_data WHERE discord_id=$1", discID).Scan(&connectedHandle)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", codeforces.ErrUserNotConnected
	}
	if err != nil {
		return "", err
	}

	return connectedHandle, nil
}
