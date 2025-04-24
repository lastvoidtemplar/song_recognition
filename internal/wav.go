package internal

import (
	"encoding/binary"
	"errors"
	"iter"
	"log/slog"
	"os"
)

type wavHeader struct {
	sampleRate uint32
	dataSize   uint32
}

type WavParser struct {
	wavPath   string
	wavHeader wavHeader

	wavFile *os.File
}

func NewWavParser(wavPath string, logger *slog.Logger) (*WavParser, error) {
	wavFile, err := os.Open(wavPath)
	if err != nil {
		slog.With(slog.String("wav_path", wavPath), slog.String("err", err.Error())).Error("Couldn`t open wav file")
		return nil, err
	}

	parser := &WavParser{
		wavPath: wavPath,
		wavFile: wavFile,
	}

	parser.parseHeader(logger)

	return parser, nil
}

func (parser *WavParser) Close() error {
	slog.With(slog.String("wav_path", parser.wavPath)).Debug("The file is closed")
	return parser.wavFile.Close()
}

// http://soundfile.sapp.org/doc/WaveFormat/
func (parser *WavParser) parseHeader(logger *slog.Logger) bool {
	logger = logger.With(slog.String("wav_path", parser.wavPath))

	buf := make([]byte, 78)
	parser.wavFile.Read(buf)

	riff := string(buf[:4])
	if riff != "RIFF" {
		logger.With(slog.String("riff", riff)).Warn("The first 4 bytes must be 'RIFF'")
		return false
	}

	format := string(buf[8:12])
	if format != "WAVE" {
		logger.With(slog.String("format", format)).Warn("The bytes between 8-12 must be 'WAVE'")
		return false
	}

	subId1 := string(buf[12:16])
	if subId1 != "fmt " {
		logger.With(slog.String("subId1", subId1)).Warn("The bytes between 12-16 must be 'fmt '")
		return false
	}

	subSize1 := binary.LittleEndian.Uint16(buf[16:20])
	audioFormat := binary.LittleEndian.Uint16(buf[20:22])
	numChannels := binary.LittleEndian.Uint16(buf[22:24])
	if subSize1 != 16 || audioFormat != 1 || numChannels != 1 {
		logger.With(
			slog.Uint64("sub_size_1", uint64(subSize1)),
			slog.Uint64("audio_format", uint64(audioFormat)),
			slog.Uint64("number_of_channels", uint64(numChannels)),
		).Warn("The file isn`t PCM")
		return false
	}

	sampleRate := binary.LittleEndian.Uint32(buf[24:28])
	//byte rate(28:32) = sample_rate * num_channels * bits_per_sample/8
	//block align (32:34) = num_channels * bits_per_sample/8
	//bitsPerSample := binary.LittleEndian.Uint16(buf[34:36])

	//LIST CHUNK between 36 and 70

	data := string(buf[70:74])
	if data != "data" {
		logger.With(slog.String("data", data)).Warn("The bytes between 70-74 must be 'data'")
		return false
	}

	dataSize := binary.LittleEndian.Uint32(buf[74:78])

	parser.wavHeader = wavHeader{
		sampleRate: sampleRate,
		dataSize:   dataSize,
	}

	logger.With(
		slog.String("wav_path", parser.wavPath),
		slog.Uint64("sample_rate", uint64(parser.wavHeader.sampleRate)),
		slog.Uint64("data_size", uint64(parser.wavHeader.dataSize)),
	).Debug("The wav header is parse successfully")

	return true
}

func (parser *WavParser) WindowsCount(windowSize int, step int) int {
	return 1 + (int(parser.wavHeader.dataSize)-2*windowSize)/(2*step)

}

var ErrTooBigWindowSize = errors.New("the window size can`t be bigger than the sample data size")

func (parser *WavParser) NewWindowIter(windowSize int, hopSize int, logger *slog.Logger) (iter.Seq2[int, []float64], error) {
	logger = logger.With(slog.String("wav_path", parser.wavPath))

	if windowSize > int(parser.wavHeader.dataSize) {
		logger.Debug("The window size can`t be bigger than the sample data size")
		return nil, ErrTooBigWindowSize
	}

	return func(yield func(int, []float64) bool) {
		numWindows := parser.WindowsCount(windowSize, hopSize)

		buf := make([]byte, 2*windowSize)
		window := make([]float64, windowSize)
		temp := make([]float64, hopSize)
		pos := 0
		for i := 0; i < numWindows; i++ {
			_, err := parser.wavFile.Read(buf[2*pos:])
			if err != nil {
				logger.With(slog.String("err", err.Error())).Warn("Error while reading the wav file")
				return
			}

			for j := pos; j < windowSize; j++ {
				sample := int16(binary.LittleEndian.Uint16(buf[2*j : 2*j+2]))
				window[j] = float64(sample)
			}

			if !yield(i, window) {
				break
			}
			pos = copy(temp, window[windowSize-hopSize:])
			pos = copy(window[:pos], temp)
		}
	}, nil
}
