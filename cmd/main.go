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
	var region string
	flag.BoolVar(&production, "prod", false, "Set environments to production")
	flag.StringVar(&region, "region", "eu-central-1", "Set the aws region")

	logger := internal.NewLogger()

	flag.Parse()

	var db internal.DB
	var err error
	if !production {
		db, err = internal.NewDBSqlite("db.sqlite", logger)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to create a DB")
			return
		}
	} else {
		awsClient, err := internal.NewAWSClient(region, logger)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to create a AWS client")
			return
		}

		options, err := awsClient.LoadMysqlDBParameters(logger)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to extracting MySql connection options")
			return
		}

		db, err = internal.NewDBMysql(options, logger)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to create a DB")
			return
		}
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
	handler = createLoggingMiddleware(handler, logger)

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
func createLoggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(fmt.Sprintf("Request Path: %s, Query Params: %s", r.URL.Path, r.URL.RawQuery))

		next.ServeHTTP(w, r)
	})
}
