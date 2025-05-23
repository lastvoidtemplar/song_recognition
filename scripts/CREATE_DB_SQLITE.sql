CREATE TABLE IF NOT EXISTS songs (
    song_id INTEGER PRIMARY KEY AUTOINCREMENT,
    song_title TEXT,
    song_url TEXT
);

CREATE  UNIQUE INDEX IF NOT EXISTS songs_song_url ON songs(song_url);

CREATE TABLE IF NOT EXISTS fingerprints (
    fingerprint_id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash_key INTEGER NOT NULL,
    song_id TEXT NOT NULL,
    song_timestamp INTEGER NOT NULL,
    FOREIGN KEY(song_id) REFERENCES songs(song_id)
);

CREATE INDEX IF NOT EXISTS fingerprints_hash_key ON fingerprints(hash_key);