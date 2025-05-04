package main

import (
	"log/slog"
	"net/http"

	"github.com/lastvoidtemplar/song_recognition/internal"
)

func main() {
	logger := internal.NewLogger()

	db, err := internal.NewDB("db.sqlite", logger)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Failed to create a DB")
		return
	}

	downloader, err := internal.NewYtDlpDownloader("downloads", logger)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Failed to create a youtube downloader")
		return
	}

	fs := http.FileServer(http.Dir("web"))

	http.HandleFunc("GET /songs", createGetSongsPaginationHandler(db, logger))
	http.HandleFunc("POST /songs", createAddSongHandler(downloader, db, logger))
	http.HandleFunc("POST /match", createMatchSongHandler("uploads", db, logger))

	http.Handle("/", http.StripPrefix("/", fs))
	http.ListenAndServe(":3000", nil)
}
