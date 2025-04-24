package internal

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func NewDB(dbPath string, logger *slog.Logger) (*DB, error) {
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
	return &DB{
		db: db,
	}, nil
}

func (db *DB) InsertSong(songUrl string, logger *slog.Logger) (int, error) {
	res, err := db.db.Exec("INSERT INTO songs (song_url) VALUES (?)",
		songUrl)

	if err != nil {
		logger.With(
			slog.String("song_url", songUrl),
			slog.String("err", err.Error()),
		).Debug("Error while inserting a song")
		return -1, err
	}

	songId, err := res.LastInsertId()

	if err != nil {
		logger.With(
			slog.String("song_url", songUrl),
			slog.String("err", err.Error()),
		).Debug("Error while inserting a song")
		return -1, err
	}

	return int(songId), nil
}

func (db *DB) InsertFingerprint(hash uint64, songId int, timestamp uint32, logger *slog.Logger) error {
	_, err := db.db.Exec("INSERT INTO fingerprints (hash_key, song_id, song_timestamp) VALUES (?, ?, ?)",
		hash, songId, timestamp)

	if err != nil {
		logger.With(
			slog.String("hash", fmt.Sprintf("%x", hash)),
			slog.String("err", err.Error()),
		).Debug("Error while inserting a fingerprint")
		return err
	}

	return nil
}

func (db *DB) GetSongUrl(songId int, logger *slog.Logger) (string, error) {
	row := db.db.QueryRow("SELECT song_url FROM songs WHERE song_id = ?", songId)

	err := row.Err()
	if err != nil {
		logger.With(
			slog.Int("song_id", songId),
			slog.String("err", err.Error()),
		).Debug("Error while inserting a fingerprint")
		return "", err
	}

	songUrl := ""
	err = row.Scan(&songUrl)

	if err != nil {
		logger.With(
			slog.Int("song_id", songId),
			slog.String("err", err.Error()),
		).Debug("Error while inserting a fingerprint")
		return "", err
	}

	return songUrl, nil
}

type Fingerprint struct {
	HashKey   uint64
	SongId    int
	Timestamp uint32
}

func (db *DB) SearchFingerprints(hashes []uint64, logger *slog.Logger) (map[uint64][]Fingerprint, error) {
	query := fmt.Sprintf("SELECT hash_key, song_id, song_timestamp FROM fingerprints WHERE hash_key IN (%s)", joinHashes(hashes))
	rows, err := db.db.Query(query)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Debug("Error while searching for fingerprints")
		return nil, err
	}
	defer rows.Close()

	matches := make(map[uint64][]Fingerprint, 0)

	for rows.Next() {
		var fingerprint Fingerprint
		err = rows.Scan(&fingerprint.HashKey, &fingerprint.SongId, &fingerprint.Timestamp)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Debug("Error while searching for fingerprints")
			return nil, err
		}
		matches[fingerprint.HashKey] = append(matches[fingerprint.HashKey], fingerprint)
	}

	if err = rows.Err(); err != nil {
		logger.With(slog.String("err", err.Error())).Debug("Error while searching for fingerprints")
		return nil, err
	}

	return matches, nil
}

func joinHashes(hashes []uint64) string {
	if len(hashes) == 0 {
		return ""
	}

	res := fmt.Sprintf("%d", hashes[0])
	for i := 1; i < len(hashes); i++ {
		res += fmt.Sprintf(", %d", hashes[i])
	}

	return res
}
