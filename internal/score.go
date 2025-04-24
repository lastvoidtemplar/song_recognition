package internal

import "math"

func ScoreFingerprints(recordingFingerprints map[uint64]uint32, dbFingerprints map[uint64][]Fingerprint) map[int]int {
	songsMatches := make(map[int][][2]uint32)

	for hash, fingerprints := range dbFingerprints {
		for _, fingerprint := range fingerprints {
			recordingTime, found := recordingFingerprints[hash]

			if found {
				songsMatches[fingerprint.SongId] = append(songsMatches[fingerprint.SongId],
					[2]uint32{recordingTime, fingerprint.Timestamp})
			}
		}
	}

	songsScores := make(map[int]int)

	for songId, matches := range songsMatches {
		score := 0

		for i := 0; i < len(matches); i++ {
			for j := i + 1; j < len(matches); j++ {
				diff1 := math.Abs(float64(matches[i][0]) - float64(matches[j][0]))
				diff2 := math.Abs(float64(matches[i][1]) - float64(matches[j][1]))
				if math.Abs(diff1-diff2) < 50 {
					score++
				}
			}
		}

		songsScores[songId] = score
	}

	return songsScores
}
