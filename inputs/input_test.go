package inputs

import (
	"testing"
	"time"

	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFileInput(t *testing.T) {

	logger.InitLogger(zapcore.DebugLevel)

	wd, _ := utils.GetCurrentWorkingDir()
	absPath := utils.ConvertToAbsolutePath("../assets/test/test.mp4", wd)

	fileInput := FileInput{}
	err := fileInput.Open(absPath)

	if err != nil {
		t.Fatalf(err.Error())
	}

	go func() {
		sum := 0.0
		for {
			packet := <-fileInput.videoOut
			if packet != nil {
				sum += float64(int(packet.duration)) / float64(fileInput.videoTimescale.Den)
				logger.Debug("VideoPacket", zap.Float64("durationSum", sum))
			}

			FreePacket(packet)
		}

	}()

	go func() {
		sum := 0.0
		for {
			packet := <-fileInput.audioOut
			if packet != nil {
				sum += float64(int(packet.duration)) / float64(fileInput.audioTimescale.Den)
				logger.Debug("AudioPacket", zap.Float64("durationSum", sum))
			}

			FreePacket(packet)

		}

	}()

	time.Sleep(time.Second * 10)
	logger.Logger.Sync()
	fileInput.Close()
}
