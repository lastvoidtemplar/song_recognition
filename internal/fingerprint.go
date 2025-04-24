package internal

import "math/cmplx"

type peakRange struct {
	min int
	max int
}

func GenerateFingerprints(spectrogram [][]complex128, timePerColumn float64) map[uint64]uint32 {
	peaksRanges := []peakRange{{40, 80}, {80, 120}, {120, 180}, {180, 300}}
	fingerprints := make(map[uint64]uint32)

	maxFreqPerRange := make([]uint64, len(peaksRanges))

	for colmInd, colm := range spectrogram {
		for peakRangeInd, peakRange := range peaksRanges {
			maxFreq := -1
			maxMag := -1.0

			for freq, num := range colm[peakRange.min:peakRange.max] {
				mag := cmplx.Abs(num)
				if mag > maxMag {
					maxFreq = peakRange.min + freq
					maxMag = mag
				}
			}

			maxFreqPerRange[peakRangeInd] = uint64(maxFreq)
		}

		h := hash(maxFreqPerRange[0], maxFreqPerRange[1], maxFreqPerRange[2], maxFreqPerRange[3])
		fingerprints[h] = uint32(float64(colmInd) * timePerColumn * 1000)
	}

	return fingerprints
}

func hash(p1, p2, p3, p4 uint64) uint64 {
	const fuzzFactor = 2

	m4 := (p4 - (p4 % fuzzFactor)) << 30
	m3 := (p3 - (p3 % fuzzFactor)) << 20
	m2 := (p2 - (p2 % fuzzFactor)) << 10
	m1 := p1 - (p1 % fuzzFactor)

	return m4 | m3 | m2 | m1
}
