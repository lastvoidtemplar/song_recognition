package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path"
	"slices"

	"github.com/google/uuid"
	"github.com/lastvoidtemplar/sabbac/internal"
)

type SongDTO struct {
	SongId string `json:"song_id"`
}

func createAddSongHandler(downloader internal.YouTubeDownloader, db *internal.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := generateReqId()
		logger = logger.With(slog.String("request_id", reqId))

		var dto SongDTO

		err := json.NewDecoder(r.Body).Decode(&dto)

		if err != nil {
			logger.Debug(err.Error())
			http.Error(w, "Invalid json format", http.StatusBadRequest)
			return
		}

		if !internal.ValidateUrl(dto.SongId) {
			logger.With(slog.String("url", dto.SongId)).Debug("Invalid song id")
			http.Error(w, "Invalid song id", http.StatusBadRequest)
			return
		}

		go func() {
			logger = logger.With(slog.String("url", dto.SongId))
			wavPath, err := downloader.DownloadWav(dto.SongId, logger)
			if err != nil {
				logger.With(slog.String("err", err.Error())).Warn("Error while downloading a song")
				return
			}

			spectrogram, timePerColm := internal.STFT(wavPath, logger)

			fingerprints := internal.GenerateFingerprints(spectrogram, timePerColm)

			songId, err := db.InsertSong(dto.SongId, logger)

			if err != nil {
				logger.With(slog.String("err", err.Error())).Warn("Error while inserting a song")
				return
			}

			logger = logger.With(slog.Int("song_id", songId))

			for hash, timestamp := range fingerprints {
				err = db.InsertFingerprint(hash, songId, timestamp, logger)

				if err != nil {
					logger.With(slog.String("err", err.Error())).Warn("Error while inserting a fingerprint")
					return
				}
			}
		}()
	}
}

func createMatchSongHandler(uploadPath string, db *internal.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := generateReqId()
		logger = logger.With(slog.String("request_id", reqId))

		r.ParseMultipartForm(10 << 20)

		audio, headers, err := r.FormFile("audio")
		if err != nil {
			logger.With(slog.String("err", err.Error())).Debug("Failed to upload the webm file")
			http.Error(w, "Failed to upload webm file", http.StatusBadRequest)
			return
		}
		defer audio.Close()

		webmPath := path.Join(uploadPath, headers.Filename)
		file, err := os.Create(webmPath)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Failed to create the webm file")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, audio)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Failed to copy the webm file")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		wavPath, err := internal.ConvertWebmToWav(webmPath, logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to convert the .webm to .wav")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		err = os.Remove(webmPath)
		if err != nil {
			logger.With(slog.String("webm_path", webmPath), slog.String("err", err.Error())).Warn("Failed to delete the .webm")
		}

		spectrogram, timePerColm := internal.STFT(wavPath, logger)
		recordingFingerprints := internal.GenerateFingerprints(spectrogram, timePerColm)

		dbFingerprints, err := db.SearchFingerprints(slices.Collect(maps.Keys(recordingFingerprints)), logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to search the database for fingerprints")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		scores := internal.ScoreFingerprints(recordingFingerprints, dbFingerprints)

		maxScore := -1
		matchSongId := -1
		for songId, score := range scores {
			if maxScore < score {
				matchSongId = songId
				maxScore = score
			}
		}

		songUrl, err := db.GetSongUrl(matchSongId, logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to get song url from song id")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(songUrl))
	}
}

func generateReqId() string {
	return uuid.NewString()
}
