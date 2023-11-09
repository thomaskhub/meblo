package outputs

import (
	"testing"
	"time"

	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap/zapcore"
)

func TestOutputBase(t *testing.T) {

	logger.InitLogger(zapcore.DebugLevel)

	wd, _ := utils.GetCurrentWorkingDir()
	absPath := utils.ConvertToAbsolutePath("../assets/test/test.mp4", wd)

	fileInput := inputs.Input{}
	fileInput.Open(absPath, false)

	output := OutputBase{}
	output.ConfigVideo(fileInput.GetVideoOut(), fileInput.GetVideoCodecParams())
	output.ConfigAudio(fileInput.GetAudioOut(), fileInput.GetAudioCodecParams())
	output.Open("/tmp/test.mp4")

	// audioParams := fileInput.GetAudioCodecParams()
	// a := fileInput.GetAudioOut()
	// output.ConfigAudio(a, audioParams)

	time.Sleep(time.Second * 10)

}
