package internal

import (
	"fmt"
	"log/slog"
)

type DB interface {
	SetupDB(logger *slog.Logger) error
	InsertSong(songTitle string, songUrl string, logger *slog.Logger) (int, error)
	InsertFingerprint(hash uint64, songId int, timestamp uint32, logger *slog.Logger) error
	GetSongsCount(logger *slog.Logger) (int, error)
	GetSongsPagination(page int, limit int, logger *slog.Logger) ([]Song, error)
	CheckSongByUrl(songUrl string, logger *slog.Logger) (bool, error)
	GetSongById(songId int, logger *slog.Logger) (Song, error)
	SearchFingerprints(hashes []uint64, logger *slog.Logger) (map[uint64][]Fingerprint, error)
}
type Song struct {
	SongId    int
	SongTitle string
	SongUrl   string
}

type Fingerprint struct {
	HashKey   uint64
	SongId    int
	Timestamp uint32
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
