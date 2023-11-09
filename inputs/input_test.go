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

	logger.InitLogger(zapcore.DebugLevel)

	wd, _ := utils.GetCurrentWorkingDir()
	absPath := utils.ConvertToAbsolutePath("../assets/test/test.mp4", wd)

	fileInput := Input{}
	err := fileInput.Open(absPath, true)

	if err != nil {
		t.Fatalf(err.Error())
	}

	videoSum := 0.0
	go func() {
		for {
			packet := <-fileInput.videoOut

			if packet != nil {
				videoSum += float64(int(packet.GetDuration())) / float64(fileInput.videoTimescale.Den)
				logger.Debug("VideoPacket", zap.Float64("durationSum", videoSum))
			}

			packet.AVPacketFree()
		}

	}()

	audioSum := 0.0
	go func() {
		for {
			packet := <-fileInput.audioOut
			if packet != nil {
				audioSum += float64(int(packet.GetDuration())) / float64(fileInput.audioTimescale.Den)
				logger.Debug("AudioPacket", zap.Float64("durationSum", audioSum))
			}

			packet.AVPacketFree()
		}

	}()

	time.Sleep(time.Second * 10)
	assert.Greater(t, videoSum, 220.0)
	assert.Greater(t, audioSum, 220.0)
	logger.Logger.Sync()
}
