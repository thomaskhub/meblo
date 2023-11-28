package inputs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFileInput(t *testing.T) {

	logger.InitLogger(zapcore.ErrorLevel)

	wd, _ := utils.GetCurrentWorkingDir()
	absPath := utils.ConvertToAbsolutePath("../assets/test/test.mp4", wd)

	fileInput := NewInputV2()
	err := fileInput.Open(absPath, true)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fileInput.Run()

	videoSum := 0.0
	dataChannel := fileInput.GetDataChannel()

	meta := fileInput.GetMetaData()

	go func() {
		for {
			packet := <-(*dataChannel)[OUT_VIDEO_CH]
			den := meta[OUT_VIDEO_CH].TimeBase.Den()

			if packet != nil {
				videoSum += float64(int(packet.Duration())) / float64(den)
				logger.Debug("VideoPacket", zap.Float64("durationSum", videoSum))
			}

			packet.Free()
		}

	}()

	audioSum := 0.0
	go func() {
		for {
			packet := <-(*dataChannel)[OUT_AUDIO_CH]
			den := meta[OUT_AUDIO_CH].TimeBase.Den()
			if packet != nil {
				audioSum += float64(int(packet.Duration())) / float64(den)
				logger.Debug("AudioPacket", zap.Float64("durationSum", audioSum))
			}

			packet.Free()
		}

	}()

	time.Sleep(time.Second * 10)
	assert.Greater(t, videoSum, 100.0)
	assert.Greater(t, audioSum, 100.0)
	logger.Logger.Sync()
}
