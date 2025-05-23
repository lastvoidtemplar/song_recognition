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
	"strconv"

	"github.com/google/uuid"
	"github.com/lastvoidtemplar/song_recognition/internal"
)

const InternalServerErrorMsg = "Internal error occured"

type ViewSongsDTO struct {
	Songs []ViewSongDTO `json:"songs"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int           `json:"total"`
}

type ViewSongDTO struct {
	SongId    int    `json:"song_id"`
	SongTitle string `json:"song_title"`
	SongUrl   string `json:"song_url"`
}

func createGetSongsPaginationHandler(db internal.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := generateReqId()
		logger := logger.With(slog.String("request_id", reqId))

		query := r.URL.Query()
		page := 1
		limit := 14

		if t := query.Get("page"); t != "" {
			i, err := strconv.Atoi(t)
			if err == nil && 0 < i {
				page = i
			}
		}

		if t := query.Get("limit"); t != "" {
			i, err := strconv.Atoi(t)
			if err == nil && 0 < i {
				limit = i
			}
		}

		count, err := db.GetSongsCount(logger)
		if err != nil {
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		songs, err := db.GetSongsPagination(page, limit, logger)
		if err != nil {
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		dto := ViewSongsDTO{
			Songs: make([]ViewSongDTO, len(songs)),
			Page:  page,
			Limit: limit,
			Total: count,
		}

		for i, song := range songs {
			dto.Songs[i].SongId = song.SongId
			dto.Songs[i].SongTitle = song.SongTitle
			dto.Songs[i].SongUrl = song.SongUrl
		}

		logger.With(
			slog.Int("page", page),
			slog.Int("limit", limit),
		).Debug("Get paginated songs successfully")

		respBody, err := json.Marshal(dto)
		if err != nil {
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Error while handling a get paginated song request")
			return
		}

		w.Write(respBody)
	}
}

type AddSongDTO struct {
	SongUrl string `json:"song_url"`
}

func createAddSongHandler(downloader internal.YouTubeDownloader, db internal.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := generateReqId()
		logger := logger.With(slog.String("request_id", reqId))

		var dto AddSongDTO

		err := json.NewDecoder(r.Body).Decode(&dto)

		if err != nil {
			logger.Debug(err.Error())
			sendError(w, "Invalid json format", http.StatusBadRequest)
			return
		}

		if !internal.ValidateUrl(dto.SongUrl) {
			logger.With(slog.String("url", dto.SongUrl)).Debug("Invalid song url")
			sendError(w, "Invalid song url", http.StatusBadRequest)
			return
		}

		url, _ := internal.StripUrl(dto.SongUrl)

		found, err := db.CheckSongByUrl(url, logger)

		if err != nil {
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			logger.With(
				slog.String("url", url),
				slog.String("err", err.Error()),
			).Warn("Error while adding a song")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		if found {
			logger.With(slog.String("url", url)).Debug("This song already exists")
			sendError(w, "This song already exists", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)

		go func() {
			logger = logger.With(slog.String("url", url))
			title, wavPath, err := downloader.DownloadWav(url, logger)
			if err != nil {
				return
			}

			spectrogram, timePerColm := internal.STFT(wavPath, logger)

			err = os.Remove(wavPath)
			if err != nil {
				logger.With(slog.String("wav_path", wavPath), slog.String("err", err.Error())).Warn("Failed to delete the .wav")
			}

			fingerprints := internal.GenerateFingerprints(spectrogram, timePerColm)

			songId, err := db.InsertSong(title, url, logger)

			if err != nil {
				return
			}

			logger = logger.With(slog.Int("song_id", songId))

			for hash, timestamp := range fingerprints {
				err = db.InsertFingerprint(hash, songId, timestamp, logger)

				if err != nil {
					return
				}
			}
		}()
	}
}

func createMatchSongHandler(uploadPath string, db internal.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := generateReqId()
		logger := logger.With(slog.String("request_id", reqId))

		r.ParseMultipartForm(10 << 20)

		audio, headers, err := r.FormFile("audio")
		if err != nil {
			logger.With(slog.String("err", err.Error())).Debug("Failed to upload the webm file")
			sendError(w, "Failed to upload webm file", http.StatusBadRequest)
			return
		}
		defer audio.Close()

		webmPath := path.Join(uploadPath, headers.Filename)
		file, err := os.Create(webmPath)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Failed to create the webm file")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, audio)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Failed to copy the webm file")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		wavPath, err := internal.ConvertWebmToWav(webmPath, logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to convert the .webm to .wav")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		err = os.Remove(webmPath)
		if err != nil {
			logger.With(slog.String("webm_path", webmPath), slog.String("err", err.Error())).Warn("Failed to delete the .webm")
		}

		spectrogram, timePerColm := internal.STFT(wavPath, logger)

		err = os.Remove(wavPath)
		if err != nil {
			logger.With(slog.String("wav_path", wavPath), slog.String("err", err.Error())).Warn("Failed to delete the .wav")
		}

		recordingFingerprints := internal.GenerateFingerprints(spectrogram, timePerColm)

		dbFingerprints, err := db.SearchFingerprints(slices.Collect(maps.Keys(recordingFingerprints)), logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to search the database for fingerprints")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
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

		song, err := db.GetSongById(matchSongId, logger)

		if err != nil {
			logger.With(slog.String("err", err.Error())).Warn("Failed to get song url from song id")
			http.Error(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}
		var dto ViewSongDTO

		dto.SongId = song.SongId
		dto.SongTitle = song.SongTitle
		dto.SongUrl = song.SongUrl

		logger.With(
			slog.Int("song_id", matchSongId),
			slog.Int("score", maxScore),
		).Debug("Match recording successfully")

		respBody, err := json.Marshal(dto)
		if err != nil {
			logger.With(
				slog.String("err", err.Error()),
			).Warn("Error while marshaling the response of the match song")
			sendError(w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		w.Write(respBody)
	}
}

func generateReqId() string {
	return uuid.NewString()
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, msg string, status int) {
	errResp := ErrorResponse{
		Error: msg,
	}

	resp, _ := json.Marshal(errResp)
	http.Error(w, string(resp), status)
}
