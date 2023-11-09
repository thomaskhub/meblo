package outputs

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
//#include <libavcodec/avcodec.h>
import "C"
import (
	"github.com/thomaskhub/meblo/ffmpeg"
	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap"
)

const (
	OutputModeVideo = "video"
	OutputModeAudio = "audio"
)

type OutputBase struct {
	ctx                   *ffmpeg.AVFormatContext
	videoCodecParamsList  []ffmpeg.AVCodecParameters
	audioCodecParamsList  []ffmpeg.AVCodecParameters
	videoOutputStreamList []*ffmpeg.AVStream
	audioOutputStreamList []*ffmpeg.AVStream
	videoIn               [](chan *ffmpeg.AVPacket)
	audioIn               [](chan *ffmpeg.AVPacket)
	outStr                string
}

func (output *OutputBase) ConfigVideo(out chan *ffmpeg.AVPacket, codecParameters ffmpeg.AVCodecParameters) {
	output.videoIn = append(output.videoIn, out)
	output.videoCodecParamsList = append(output.videoCodecParamsList, codecParameters)
}

func (output *OutputBase) ConfigAudio(out chan *ffmpeg.AVPacket, codecParameters ffmpeg.AVCodecParameters) {
	output.audioIn = append(output.audioIn, out)
	output.audioCodecParamsList = append(output.audioCodecParamsList, codecParameters)
}

func (output *OutputBase) Open(outStr string) {
	output.outStr = outStr
	output.ctx = &ffmpeg.AVFormatContext{}

	//open ffmpeg output context 2
	ret := output.ctx.AVFormatAllocOutputContext2(outStr)
	if ret < 0 {
		logger.Fatal("Output", zap.String("OutputFormatError", outStr))
	}

	//create the streams for video
	for i, codecParams := range output.videoCodecParamsList {
		output.videoOutputStreamList = append(output.videoOutputStreamList, output.ctx.AVFormatNewStream())
		output.videoOutputStreamList[i].SetCodecParameters(codecParams)
	}

	for i, codecParams := range output.videoCodecParamsList {
		output.videoOutputStreamList = append(output.videoOutputStreamList, output.ctx.AVFormatNewStream())
		output.videoOutputStreamList[i].SetCodecParameters(codecParams)
	}

	output.ctx.CheckIfNoFile(outStr)
	output.ctx.AVFormatWriteHeader()
}

func (output *OutputBase) Close() {
	output.ctx.AVFormatFreeContext()
}
