package internal

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

type DBSMySql struct {
	db *sql.DB
}

type DBMySqlConnectionOption struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func NewDBMysql(options DBMySqlConnectionOption, logger *slog.Logger) (DB, error) {
	loggerNew := logger.With(slog.String("db_host", options.Host), slog.String("db_name", options.DBName))

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", options.Username, options.Password, options.Host, options.Port, options.DBName)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		loggerNew.With(slog.String("err", err.Error())).Error("Error while creating a db")
		return nil, err
	}

	loggerNew.Info("DB is created successfully")
	return &DBSMySql{
		db: db,
	}, nil
}

func (db *DBSMySql) SetupDB(logger *slog.Logger) error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS songs (
    song_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    song_title VARCHAR(512),
    song_url VARCHAR(512)
	);`)

	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Error("Error while initing the songs table")
		return err
	}

	_, err = db.db.Exec(`CREATE TABLE IF NOT EXISTS fingerprints (
    fingerprint_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    hash_key INTEGER NOT NULL,
    song_id INTEGER NOT NULL,
    song_timestamp INTEGER NOT NULL,
    FOREIGN KEY(song_id) REFERENCES songs(song_id)
	);`)

	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Error("Error while initing the fingerprints table")
		return err
	}

	var existsSongsSongUrlIndex int
	checkSongsSongUrlIndexQuery := `
			SELECT COUNT(1)
			FROM information_schema.statistics
			WHERE table_schema = DATABASE()
			  AND table_name = "songs"
			  AND index_name = "songs_song_url_index"`

	err = db.db.QueryRow(checkSongsSongUrlIndexQuery).Scan(&existsSongsSongUrlIndex)
	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Error("Error while checking for song url unique index")
		return err
	}

	if existsSongsSongUrlIndex == 0 {
		createSongsUrlIndexQuery := "CREATE UNIQUE INDEX songs_song_url_index ON songs(song_url)"
		_, err = db.db.Exec(createSongsUrlIndexQuery)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Error("Error while initting the song url unique index")
			return err
		}
	}

	var existsFingerprintsHashKeyIndex int
	checkFingerprintsHashKeyIndexQuery := `
			SELECT COUNT(1)
			FROM information_schema.statistics
			WHERE table_schema = DATABASE()
			  AND table_name = "fingerprints"
			  AND index_name = "fingerprints_hash_key_index"`

	err = db.db.QueryRow(checkFingerprintsHashKeyIndexQuery).Scan(&existsFingerprintsHashKeyIndex)
	if err != nil {
		logger.With(
			slog.String("err", err.Error()),
		).Error("Error while checking for fingerprints hash key index")
		return err
	}

	if existsFingerprintsHashKeyIndex == 0 {
		createFingerprintsHashKeyIndexQuery := "CREATE UNIQUE INDEX fingerprints_hash_key_index ON fingerprints(hash_key)"
		_, err = db.db.Exec(createFingerprintsHashKeyIndexQuery)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Error("Error while initting the fingerprints hash key index")
			return err
		}
	}

	logger.Info("Db setup successfully")

	return err
}

func (db *DBSMySql) InsertSong(songTitle string, songUrl string, logger *slog.Logger) (int, error) {
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

func (db *DBSMySql) InsertFingerprint(hash uint64, songId int, timestamp uint32, logger *slog.Logger) error {
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

func (db *DBSMySql) GetSongsCount(logger *slog.Logger) (int, error) {
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

func (db *DBSMySql) GetSongsPagination(page int, limit int, logger *slog.Logger) ([]Song, error) {
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

func (db *DBSMySql) CheckSongByUrl(songUrl string, logger *slog.Logger) (bool, error) {
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

func (db *DBSMySql) GetSongById(songId int, logger *slog.Logger) (Song, error) {
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

func (db *DBSMySql) SearchFingerprints(hashes []uint64, logger *slog.Logger) (map[uint64][]Fingerprint, error) {
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
