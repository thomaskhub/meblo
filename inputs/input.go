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
	"time"

	"github.com/thomaskhub/meblo/ffmpeg"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type InputOptions struct {
	AutoRetry bool //enable auto retry
}

type Input struct {
	base           InputBase
	videoOut       chan *ffmpeg.AVPacket
	audioOut       chan *ffmpeg.AVPacket
	audioTimescale ffmpeg.Timescale
	videoTimescale ffmpeg.Timescale
	inStr          string //input string passed to open
	Options        InputOptions
}

func (file *Input) GetVideoOut() chan *ffmpeg.AVPacket {
	return file.videoOut
}

func (file *Input) GetAudioOut() chan *ffmpeg.AVPacket {
	return file.audioOut
}

func (file *Input) GetVideoCodecParams() ffmpeg.AVCodecParameters {
	return file.base.GetVideoCodecParams()
}

func (file *Input) GetAudioCodecParams() ffmpeg.AVCodecParameters {
	return file.base.GetAudioCodecParams()
}

// Close closes the input file.
//
// There are no parameters.
// It does not return anything.
func (file *Input) Close() {
	// Close the input format context
	file.base.ctx.AVFormatCloseInput()
	// C.avformat_close_input(&file.base.ctx.C)

	// Free the format context
	file.base.ctx.AVFormatFreeContext()

	// C.avformat_free_context(file.base.ctx)

	// Clear the video streams and audio streams
	file.base.videoStreams = nil
	file.base.audioStreams = nil

	// Close the video output channel
	close(file.videoOut)

	// Close the audio output channel
	close(file.audioOut)
}
func (file *Input) Open(inStr string, autoRetry bool) error {
	file.inStr = inStr
	file.Options = InputOptions{
		AutoRetry: autoRetry,
	}
	file.base = InputBase{}
	file.base.OpenInput(inStr)

	//check if we have the specified number of video streams and audio streams
	//we only take the first audio and first video stream in the file
	err := file.base.CheckHasStreams(1, 1)

	if err != nil {
		logger.Error("FileInput", zap.String("StramCheck", "failed"))
		return err
	}

	file.videoTimescale = file.base.videoStreams[0].GetTimebase()
	file.audioTimescale = file.base.audioStreams[0].GetTimebase()

	file.videoOut = make(chan *ffmpeg.AVPacket, 16)
	file.audioOut = make(chan *ffmpeg.AVPacket, 16)

	//create a loop in a go routine that will read the input data until file is close and
	//write the data to the videoOut and audioOut channels
	go func() {

		for {
			// Read the input data until the file is closed
			// and write the data to the videoOut and audioOut channels
			err := file.readPacket()
			if err == utils.FFmpegErrorEinval {
				logger.Fatal("could not read packet. abort")
			} else if err == utils.FFmpegErrorEof {
				logger.Debug("reached the end of the file")
				if file.Options.AutoRetry {
					err := file.restart()
					if err != nil {
						logger.Error("Input:", zap.String("ReadPacketError", "could not restart the file. abort"))
					}
				}
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
func (file *Input) readPacket() int {
	packet := ffmpeg.AVPacket{}
	packet.AVPacketAllocate()

	// packet := C.av_packet_alloc()
	ret := file.base.ctx.AVReadFrame(&packet)
	// ret := C.av_read_frame(formatContext, packet)

	if ret < 0 {
		// C.av_packet_free(&packet)
		packet.AVPacketFree()
		return int(ret)
	}

	if packet.GetStreamIndex() == file.base.videoStreams[0].GetStreamIndex() {
		file.videoOut <- &packet
		logger.Debug("FileInput", zap.String("VideoPacket", "written"))

	} else if packet.GetStreamIndex() == file.base.audioStreams[0].GetStreamIndex() {
		file.audioOut <- &packet
		logger.Debug("FileInput", zap.String("AudioPacket", "written"))
	}

	return 0
}

// Restart restarts the input file with a new string.
//
// Parameters:
//   - inStr: the new input string
func (file *Input) restart() error {
	file.Close()
	time.Sleep(4 * time.Second)
	return file.Open(file.inStr, file.Options.AutoRetry)
}
