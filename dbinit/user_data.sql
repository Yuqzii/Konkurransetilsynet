CREATE TABLE IF NOT EXISTS user_data (
	discord_id NUMERIC(20) NOT NULL,
	codeforces_handle VARCHAR(255)
);

ALTER TABLE user_data
ADD CONSTRAINT unique_discord_id UNIQUE(discord_id);
