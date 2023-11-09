package inputs

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
//#include <libavcodec/avcodec.h>
import "C"
import (
	"fmt"
	"log"
	"os"

	"github.com/thomaskhub/meblo/ffmpeg"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type InputBase struct {
	ctx          *ffmpeg.AVFormatContext
	videoStreams []*ffmpeg.AVStream
	audioStreams []*ffmpeg.AVStream
}

func (input *InputBase) GetVideoCodecParams() ffmpeg.AVCodecParameters {
	return input.videoStreams[0].GetCodecPar()
}

func (input *InputBase) GetAudioCodecParams() ffmpeg.AVCodecParameters {
	return input.audioStreams[0].GetCodecPar()
}

// OpenInputs() open a input. Right now supports any ffmpeg input format string
// currently we only support inputs with a single audio and / or video stream
//
// Parameters :
// - in: the ffmpeg formated input string (currentyl only ffmpeg is used for inputs)
//
// Examples:
//
// OpenInput("udp://localhost:1234")
// OpenInput("/home/user/Videos/test.mp4")

// CheckHasStreams checks if we have the specified number of video streams and audio streams.
//
// Parameters:
//
//	noVideoStreams - the specified number of video streams.
//	noAudioStreams - the specified number of audio streams.
//
// Return:
//
//	error - if the number of video or audio streams does not match the specified count.
func (input *InputBase) CheckHasStreams(noVideoStreams, noAudioStreams int) error {
	//check if we have the specified number of video streams and audio streams, if not return an error otherwise return nil
	if len(input.videoStreams) < int(noVideoStreams) || len(input.audioStreams) < int(noAudioStreams) {
		//also print the values of the streams
		logger.Debug("CheckStreams: ", zap.Int("No Video Streams", len(input.videoStreams)))
		logger.Debug("CheckStreams: ", zap.Int("No Audio Streams", len(input.audioStreams)))
		return fmt.Errorf("number of video or audio streams does not match the specified count")
	}
	return nil
}

// DumpInfo dumps the information about the input format.
func (input *InputBase) DumpInfo() {
	input.ctx.AVDumpFormat()
	// C.av_dump_format(input.formatContext, 0, nil, 0)
}

// OpenInput opens an input based on the given input string.
//
// Parameters:
//
//	inStr - the input string to open
//
// The input string specifies the input source to be opened.
// It returns an error if the input could not be opened.
func (input *InputBase) OpenInput(inStr string) error {

	isFile := utils.IsFilePath(inStr)

	//check if the input is a file
	if isFile {
		if _, err := os.Stat(inStr); os.IsNotExist(err) {
			log.Println(inStr)
			log.Fatalf("file does not exist")
			return err
		}
	}

	input.audioStreams = make([]*ffmpeg.AVStream, 0)
	input.videoStreams = make([]*ffmpeg.AVStream, 0)

	input.ctx = &ffmpeg.AVFormatContext{}
	input.ctx.AVFormatAllocContext()

	ret := input.ctx.AVFormatOpenInput(inStr)

	if ret < 0 {
		input.ctx.AVFormatFreeContext()
		return fmt.Errorf("could not open input")
	}

	ret = input.ctx.AVFormatFindStreamInfo()
	if ret < 0 {
		input.ctx.AVFormatFreeContext()
		return fmt.Errorf("could not find stream info")
	}

	for i := 0; i < input.ctx.GetNumberOfStreams(); i++ {
		stream := input.ctx.GetStreams()[i]
		codecType := stream.GetCodecType()

		if codecType == C.AVMEDIA_TYPE_VIDEO {
			input.videoStreams = append(input.videoStreams, stream)
		} else if codecType == C.AVMEDIA_TYPE_AUDIO {
			input.audioStreams = append(input.audioStreams, stream)
		}
	}

	return nil
}
