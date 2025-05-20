package internal

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
)

type DBSqlite struct {
	db *sql.DB
}

func NewDBSqlite(dbPath string, logger *slog.Logger) (DB, error) {
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		logger.With(slog.String("db_path", dbPath), slog.String("err", err.Error())).Error("Error while creating a db")
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		logger.With(slog.String("db_path", dbPath), slog.String("err", err.Error())).Error("Error while creating a db")
		return nil, err
	}

	logger.With(slog.String("db_path", dbPath)).Info("DB is created successfully")
	return &DBSqlite{
		db: db,
	}, nil
}

func (db *DBSqlite) SetupDB(logger *slog.Logger) error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS songs (
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

CREATE INDEX IF NOT EXISTS fingerprints_hash_key ON fingerprints(hash_key);`)

	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Error("Error while setuping the db")
		return err
	}

	logger.Info("Db setup successfully")

	return err
}

func (db *DBSqlite) InsertSong(songTitle string, songUrl string, logger *slog.Logger) (int, error) {
	res, err := db.db.Exec("INSERT INTO songs (song_title, song_url) VALUES (?, ?)",
		songTitle, songUrl)

	if err != nil {
		logger.With(
			slog.String("song_title", songTitle),
			slog.String("song_url", songUrl),
			slog.String("err", err.Error()),
		).Warn("Error while inserting a song")
		return -1, err
	}

	songId, err := res.LastInsertId()

	if err != nil {
		logger.With(
			slog.String("song_title", songTitle),
			slog.String("song_url", songUrl),
			slog.String("err", err.Error()),
		).Warn("Error while inserting a song")
		return -1, err
	}

	logger.With(
		slog.Int("song_id", int(songId)),
		slog.String("song_title", songTitle),
	).Debug("Song was inserted successfully")

	return int(songId), nil
}

func (db *DBSqlite) InsertFingerprint(hash uint64, songId int, timestamp uint32, logger *slog.Logger) error {
	_, err := db.db.Exec("INSERT INTO fingerprints (hash_key, song_id, song_timestamp) VALUES (?, ?, ?)",
		hash, songId, timestamp)

	if err != nil {
		logger.With(
			slog.String("hash", fmt.Sprintf("%x", hash)),
			slog.Int("song_id", songId),
			slog.String("err", err.Error()),
		).Warn("Error while inserting a fingerprint")
		return err
	}

	// logger.With(
	// 	slog.String("hash", fmt.Sprintf("%x", hash)),
	// 	slog.Int("song_id", songId),
	// ).Debug("FIngerprint was inserted successfully")

	return nil
}

func (db *DBSqlite) GetSongsCount(logger *slog.Logger) (int, error) {
	row := db.db.QueryRow("SELECT COUNT(song_id) FROM songs")

	err := row.Err()
	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Warn("Error while getting songs count")
		return 0, err
	}

	var count int
	err = row.Scan(&count)

	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Warn("Error while getting songs count")
		return 0, err
	}

	logger.With(
		slog.Int("songs_count", count),
	).Debug("Songs count was got successfully")

	return count, nil
}

func (db *DBSqlite) GetSongsPagination(page int, limit int, logger *slog.Logger) ([]Song, error) {
	rows, err := db.db.Query("SELECT song_id, song_title, song_url FROM songs LIMIT ? OFFSET ?", limit, (page-1)*limit)

	if err != nil {
		logger.With(
			slog.Int("page", page),
			slog.Int("limit", limit),
			slog.String("err", err.Error()),
		).Warn("Error while getting songs with pagination")
		return nil, err
	}

	songs := make([]Song, 0)
	for rows.Next() {
		var song Song
		err := rows.Scan(&song.SongId, &song.SongTitle, &song.SongUrl)

		if err != nil {
			logger.With(
				slog.Int("page", page),
				slog.Int("limit", limit),
				slog.String("err", err.Error()),
			).Warn("Error while getting songs with pagination")
			return nil, err
		}
		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		logger.With(
			slog.Int("page", page),
			slog.Int("limit", limit),
			slog.String("err", err.Error()),
		).Warn("Error while getting songs with pagination")
		return nil, err
	}

	logger.With(
		slog.Int("page", page),
		slog.Int("limit", limit),
	).Debug("Songs was paginated successfully")

	return songs, nil
}

func (db *DBSqlite) CheckSongByUrl(songUrl string, logger *slog.Logger) (bool, error) {
	row := db.db.QueryRow("SELECT 1 FROM songs WHERE song_url = ?", songUrl)

	err := row.Err()
	if err != nil {
		logger.With(
			slog.String("song_url", songUrl),
			slog.String("err", err.Error()),
		).Warn("Error while checking for song by song_url")
		return false, err
	}

	var found bool
	err = row.Scan(&found)

	if err != nil {
		return false, nil
	}

	return found, nil
}

func (db *DBSqlite) GetSongById(songId int, logger *slog.Logger) (Song, error) {
	var song Song
	row := db.db.QueryRow("SELECT song_id, song_title, song_url FROM songs WHERE song_id = ?", songId)

	err := row.Err()
	if err != nil {
		logger.With(
			slog.Int("song_id", songId),
			slog.String("err", err.Error()),
		).Warn("Error while getting song by song_id")
		return song, err
	}

	err = row.Scan(&song.SongId, &song.SongTitle, &song.SongUrl)

	if err != nil {
		logger.With(
			slog.Int("song_id", songId),
			slog.String("err", err.Error()),
		).Warn("Error while getting a song by song_id")
		return song, err
	}

	return song, nil
}

func (db *DBSqlite) SearchFingerprints(hashes []uint64, logger *slog.Logger) (map[uint64][]Fingerprint, error) {
	joined := joinHashes(hashes)
	query := fmt.Sprintf("SELECT hash_key, song_id, song_timestamp FROM fingerprints WHERE hash_key IN (%s)", joined)
	rows, err := db.db.Query(query)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Warn("Error while searching for fingerprints")
		return nil, err
	}
	defer rows.Close()

	matches := make(map[uint64][]Fingerprint, 0)

	for rows.Next() {
		var fingerprint Fingerprint
		err = rows.Scan(&fingerprint.HashKey, &fingerprint.SongId, &fingerprint.Timestamp)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Error while searching for fingerprints")
			return nil, err
		}
		matches[fingerprint.HashKey] = append(matches[fingerprint.HashKey], fingerprint)
	}

	if err = rows.Err(); err != nil {
		logger.With(slog.String("err", err.Error())).Warn("Error while searching for fingerprints")
		return nil, err
	}

	return matches, nil
}
