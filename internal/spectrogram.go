package internal

import (
	"image"
	"image/color"
	"image/jpeg"
	"log/slog"
	"math"
	"math/cmplx"
	"os"
)

// Spectogram slice of frequencies for window
func STFT(wavPath string, logger *slog.Logger) ([][]complex128, float64) {
	wavParser, err := NewWavParser(wavPath, logger)
	if err != nil {
		logger.With(slog.String("wav_path", wavPath)).Debug("Couldn`t create the wav parser")
		return nil, 0
	}
	defer wavParser.Close()

	const windowSize = 1024
	const hopSize = 512

	numWindows := wavParser.WindowsCount(windowSize, hopSize)
	stftRes := make([][]complex128, numWindows)
	windowFunction := hammingWindow(windowSize)

	iter, err := wavParser.NewWindowIter(windowSize, hopSize, logger)

	if err != nil {
		logger.With(slog.String("wav_path", wavPath)).Debug("Couldn`t create an iterator for the sample data")
		return nil, 0
	}

	for i, sample := range iter {
		applyWindowFunction(sample, windowFunction)
		stftRes[i] = fft(sample, 0, len(sample)-1, 1)[:windowSize/2]
	}

	timePerColumn := float64(hopSize) / float64(wavParser.wavHeader.sampleRate)

	return stftRes, timePerColumn
}

// https://en.wikipedia.org/wiki/Window_function#Hann_and_Hamming_windows
func hammingWindow(windowSize int) []float64 {
	const a0 float64 = 25.0 / 46
	const a1 = 1 - a0
	const piTimes2 float64 = 2 * math.Pi
	var windowSizeFloat64 float64 = float64(windowSize)

	window := make([]float64, windowSize)
	for i := 0; i < windowSize; i++ {
		window[i] = a0 - a1*math.Cos(piTimes2*float64(i)/windowSizeFloat64)
	}

	return window
}

func applyWindowFunction(sample []float64, windowFunction []float64) {
	for i, v := range windowFunction {
		sample[i] *= v
	}
}

// this is here for testing purposes only
// the fft is much faster
func dft(sample []float64) []float64 {
	frequencies := make([]float64, 1+len(sample)/2)

	for fr := 0; fr < len(frequencies); fr++ {
		w := cmplx.Exp(complex(0, -2*float64(math.Pi)*float64(fr)/float64(len(sample))))
		e := complex(1.0, 0)
		sum := complex(0, 0)
		for j := 0; j < len(sample); j++ {
			sum += complex(sample[j], 0) * e
			e *= w
		}
		frequencies[fr] = cmplx.Abs(sum)
	}

	return frequencies
}

func fft(sample []float64, st int, end int, step int) []complex128 {
	n := (end-st)/(step) + 1

	if n == 1 {
		return []complex128{complex(sample[st], 0)}
	}

	frequencies := make([]complex128, n)

	frequenciesEven := fft(sample, st, end-step, 2*step)
	frequenciesOdd := fft(sample, st+step, end, 2*step)

	w := cmplx.Exp(complex(0, -2*float64(math.Pi)/float64(n)))
	e := complex(1, 0)
	for i := 0; i < n/2; i++ {
		frequencies[i] = frequenciesEven[i] + e*frequenciesOdd[i]
		frequencies[i+n/2] = frequenciesEven[i] - e*frequenciesOdd[i]
		e *= w
	}

	return frequencies
}

// this function is for testing purposes
func SpectrogramToImage(path string, spectrogram [][]complex128, logger *slog.Logger) error {
	img := image.NewRGBA(image.Rect(0, 0, len(spectrogram), len(spectrogram[0])))

	maxMag := -1.0
	for _, bin := range spectrogram {
		for _, c := range bin {
			mag := cmplx.Abs(c)
			if mag > maxMag {
				maxMag = mag
			}
		}
	}

	con := 255 / maxMag

	for binInd, bin := range spectrogram {
		maxFr := len(bin)
		for freqInd, c := range bin {
			grey := uint8(cmplx.Abs(c) * con)
			img.Set(binInd, maxFr-freqInd, color.RGBA{R: grey, G: grey, B: grey, A: 255})
		}
	}

	imgFile, err := os.Create(path)

	if err != nil {
		logger.With(slog.String("path", path), slog.String("err", err.Error())).Warn("Couldn`t create the image")
		return err
	}

	jpeg.Encode(imgFile, img, nil)
	imgFile.Close()

	return nil
}
