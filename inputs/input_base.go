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
	"unsafe"

	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type BaseInput struct {
	formatContext *C.AVFormatContext
	videoStreams  []*C.AVStream
	audioStreams  []*C.AVStream
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
func (input *BaseInput) CheckHasStreams(noVideoStreams, noAudioStreams int) error {
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
func (input *BaseInput) DumpInfo() {
	C.av_dump_format(input.formatContext, 0, nil, 0)
}

// OpenInput opens an input based on the given input string.
//
// Parameters:
//
//	inStr - the input string to open
//
// The input string specifies the input source to be opened.
// It returns an error if the input could not be opened.
func (input *BaseInput) OpenInput(inStr string) error {

	//check if the string is a linux windows or mac file path
	isFile := utils.IsFilePath(inStr)

	if isFile {
		if _, err := os.Stat(inStr); os.IsNotExist(err) {
			log.Println(inStr)
			log.Fatalf("file does not exist")
			return err
		}
	}

	input.audioStreams = make([]*C.AVStream, 0)
	input.videoStreams = make([]*C.AVStream, 0)

	in := C.CString(inStr)
	input.formatContext = C.avformat_alloc_context()

	ret := C.avformat_open_input(&input.formatContext, in, nil, nil)
	if ret < 0 {
		C.avformat_free_context(input.formatContext)
		return fmt.Errorf("could not open input")
	}

	ret = C.avformat_find_stream_info(input.formatContext, nil)
	if ret < 0 {
		C.avformat_free_context(input.formatContext)
		return fmt.Errorf("could not find stream info")
	}

	//extract video streams and audio streams into separate lists in the input
	numberOfStreams := int(input.formatContext.nb_streams)
	for i := 0; i < numberOfStreams; i++ {
		streams := input.formatContext.streams
		arrayOfPointers := (*[1 << 30]*C.AVStream)(unsafe.Pointer(streams))[:numberOfStreams:numberOfStreams]
		stream := arrayOfPointers[i]

		if stream.codecpar.codec_type == C.AVMEDIA_TYPE_VIDEO {
			input.videoStreams = append(input.videoStreams, stream)
		} else if stream.codecpar.codec_type == C.AVMEDIA_TYPE_AUDIO {
			input.audioStreams = append(input.audioStreams, stream)
		}
	}

	return nil
}

// freePacket frees the memory allocated for the AVPacket.
//
// packet: A pointer to the AVPacket to be freed.
func FreePacket(packet *C.AVPacket) {
	if packet != nil {
		logger.Debug("free packet")
		C.av_packet_free(&packet)
	}
}
