package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lastvoidtemplar/song_recognition/internal"
)

func main() {
	var production bool
	flag.BoolVar(&production, "env", false, "Set environments to production")

	logger := internal.NewLogger()

	flag.Parse()

	db, err := internal.NewDB("db.sqlite", logger)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Failed to create a DB")
		return
	}

	err = db.SetupDB(logger)

	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Failed to setup a DB")
		return
	}

	downloader, err := internal.NewYtDlpDownloader("downloads", logger)
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Failed to create a youtube downloader")
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /songs", createGetSongsPaginationHandler(db, logger))
	mux.HandleFunc("POST /songs", createAddSongHandler(downloader, db, logger))
	mux.HandleFunc("POST /match", createMatchSongHandler("uploads", db, logger))

	logger.Debug(fmt.Sprint(production))

	if !production {
		fs := http.FileServer(http.Dir("web"))
		mux.Handle("/", http.StripPrefix("/", fs))
	}

	handler := withCORS(mux)

	if !production {
		http.ListenAndServe(":3000", handler)
	} else {
		err := http.ListenAndServe(":80", handler)
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
