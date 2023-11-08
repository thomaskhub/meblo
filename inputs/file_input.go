package inputs

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
//#include <libavcodec/avcodec.h>
// #include <libavutil/frame.h>
// #include <libavutil/log.h>
// #include <libavutil/samplefmt.h>
// #include <errno.h>
import "C"
import (
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type FileInput struct {
	base           BaseInput
	videoOut       chan *C.AVPacket
	audioOut       chan *C.AVPacket
	audioTimescale utils.Timescale
	videoTimescale utils.Timescale
}

// create a fucntion that closes the input
func (file *FileInput) Close() {
	// Close the input format context
	C.avformat_close_input(&file.base.formatContext)

	// Free the format context
	C.avformat_free_context(file.base.formatContext)

	// Clear the video streams and audio streams
	file.base.videoStreams = nil
	file.base.audioStreams = nil

	// Close the video output channel
	close(file.videoOut)

	// Close the audio output channel
	close(file.audioOut)
}
func (file *FileInput) Open(inStr string) error {

	file.base = BaseInput{}
	file.base.OpenInput(inStr)

	//check if we have the specified number of video streams and audio streams
	//we only take the first audio and first video stream in the file
	err := file.base.CheckHasStreams(1, 1)

	if err != nil {
		logger.Error("FileInput", zap.String("StramCheck", "failed"))
		return err
	}

	file.videoTimescale.Den = int(file.base.videoStreams[0].time_base.den)
	file.videoTimescale.Num = int(file.base.videoStreams[0].time_base.num)

	file.audioTimescale.Den = int(file.base.audioStreams[0].time_base.den)
	file.videoTimescale.Num = int(file.base.audioStreams[0].time_base.num)

	file.videoOut = make(chan *C.AVPacket, 16)
	file.audioOut = make(chan *C.AVPacket, 16)

	//create a loop in a go routine that will read the input data until file is close and
	//write the data to the videoOut and audioOut channels
	go func() {

		for {
			// Read the input data until the file is closed
			// and write the data to the videoOut and audioOut channels
			err := file.readPacket(file.base.formatContext)
			if err == utils.FFmpegErrorEinval {
				logger.Fatal("could not read packet. abort")
			} else if err == utils.FFmpegErrorEof {
				logger.Debug("reached the end of the file")
				return
			} else {
				logger.Debug("test", zap.Int("ret", err))
				continue
			}

		}
	}()
	return nil
}

// readPacket reads a packet from the AVFormatContext.
//
// Parameters:
//
//	formatContext - a pointer to an AVFormatContext
//
// Returns:
//
//	packet - a pointer to an AVPacket
//	error - an error object
func (file *FileInput) readPacket(formatContext *C.AVFormatContext) int {
	if file.base.formatContext == nil {
		return -C.EINVAL
	}

	packet := C.av_packet_alloc()
	ret := C.av_read_frame(formatContext, packet)

	if ret < 0 {
		C.av_packet_free(&packet)
		return int(ret)
	}

	if packet.stream_index == file.base.videoStreams[0].index {
		file.videoOut <- packet
		logger.Debug("FileInput", zap.String("VideoPacket", "written"))
	} else if packet.stream_index == file.base.audioStreams[0].index {
		file.audioOut <- packet
		logger.Debug("FileInput", zap.String("AudioPacket", "written"))
	}

	return 0
}
