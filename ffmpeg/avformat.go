package ffmpeg

//#cgo pkg-config: libavformat libavfilter libavutil libavcodec
//#include <libavformat/avformat.h>
//#include <errno.h>
//static AVStream *getStream(AVFormatContext *ctx, int index) {
//	return ctx->streams[index];
//}
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap"
)

const (
	AVMEDIA_TYPE_VIDEO int = C.AVMEDIA_TYPE_VIDEO
	AVMEDIA_TYPE_AUDIO int = C.AVMEDIA_TYPE_AUDIO
)

// AVFormatContext represents an input or output format context.
type AVFormatContext struct {
	CAVFormatContext *C.AVFormatContext
}

func (ctx *AVFormatContext) Version() {
	fmt.Println(int(C.LIBAVFORMAT_VERSION_MAJOR), int(C.LIBAVFORMAT_VERSION_MINOR), int(C.LIBAVFORMAT_VERSION_MICRO))
}

// GetNumberOfStreams returns the number of streams in the AVFormatContext.
//
// No parameters.
// Returns an integer.
func (ctx *AVFormatContext) GetNumberOfStreams() int {
	return int(ctx.CAVFormatContext.nb_streams)
}

// GetStreams returns an array of AVStreams.
//
// It does not take any parameters.
// It returns a slice of pointers to AVStream.
// func (ctx *AVFormatContext) GetStreams() []*AVStream {
// 	streams := make([]*AVStream, ctx.GetNumberOfStreams())
// 	numberOfStreams := ctx.GetNumberOfStreams()
// 	cStrams := (*C.AVStream)(unsafe.Pointer(ctx.CAVFormatContext.streams))

// 	for i := 0; i < numberOfStreams; i++ {

// 		arrayOfPointers := (*[1 << 30]*C.AVStream)(unsafe.Pointer(cStrams))[:numberOfStreams:numberOfStreams]
// 		stream := arrayOfPointers[i]
// 		streams[i] = &AVStream{CAVStream: stream}
// 	}

// 	return streams
// }

func (ctx *AVFormatContext) GetStream(index int) *AVStream {
	cStream := C.getStream(ctx.CAVFormatContext, C.int(index))
	return &AVStream{CAVStream: (*C.AVStream)(unsafe.Pointer(cStream))}
	return nil
}

// AVFormatAllocContext allocates an AVFormatContext struct.
//
// It returns a pointer to the allocated AVFormatContext struct.
func (ctx *AVFormatContext) AVFormatAllocContext() {
	cCtx := C.avformat_alloc_context()
	ctx = &AVFormatContext{CAVFormatContext: (*C.AVFormatContext)(unsafe.Pointer(cCtx))}
}

// AVFormatAllocOutputContext2 allocates an AVFormatContext for an output format.
func (ctx *AVFormatContext) AVFormatAllocOutputContext2(url string) int {
	//call the avformat_alloc_output_context2
	urlCsting := C.CString(url)
	defer C.free(unsafe.Pointer(urlCsting))

	// formatCstr := C.CString(utils.GetOutputFormat(url))
	formatCstr := C.CString("mpegts") //TODO: this must be set based on the opoutut
	defer C.free(unsafe.Pointer(formatCstr))

	var cCtx *C.AVFormatContext
	ret := int(C.avformat_alloc_output_context2(&cCtx, nil, formatCstr, urlCsting))
	ctx.CAVFormatContext = (*C.AVFormatContext)(unsafe.Pointer(cCtx))
	return ret
}

// AVFormatNewStream creates a new stream for the AVFormatContext.
func (ctx *AVFormatContext) AVFormatNewStream() *AVStream {
	unsafePointer := unsafe.Pointer(C.avformat_new_stream(ctx.CAVFormatContext, nil))
	return &AVStream{CAVStream: (*C.AVStream)(unsafePointer)}
}

// AVFormatOpenInput opens an input stream and initializes the format context.
//
// url: The input URL.
// Returns: The status code indicating success or failure.
func (ctx *AVFormatContext) AVFormatOpenInput(url string) int {
	urlC := C.CString(url)
	defer C.free(unsafe.Pointer(urlC))
	return int(C.avformat_open_input(&ctx.CAVFormatContext, urlC, nil, nil))
}

// AVFormatFindStreamInfo finds stream information in the given AVFormatContext.
//
// It returns an integer indicating the result of the stream information finding process.
func (ctx *AVFormatContext) AVFormatFindStreamInfo() int {
	return int(C.avformat_find_stream_info(ctx.CAVFormatContext, nil))
}

// AVReadFrame reads a frame from the AVFormatContext and stores it in the AVPacket.
//
// ctx - The AVFormatContext to read from.
// packet - The AVPacket to store the frame in.
// Returns: The result of av_read_frame as an int.
func (ctx *AVFormatContext) AVReadFrame(packet *AVPacket) int {
	return int(C.av_read_frame(ctx.CAVFormatContext, packet.CAVPacket))
}

// AVWriteFrame writes a frame to the AVFormatContext.
//
// It takes no parameters.
// It returns an integer.
func (ctx *AVFormatContext) AVWriteFrame() int {
	return int(C.av_write_frame(ctx.CAVFormatContext, nil))
}

// AVFormatFreeContext frees the AVFormatContext and associated resources.
//
// ctx - The AVFormatContext to be freed.
func (ctx *AVFormatContext) AVFormatFreeContext() {
	C.avformat_free_context(ctx.CAVFormatContext)
}

// AVDumpFormat dumps the format-specific information about the given AVFormatContext.
func (ctx *AVFormatContext) AVDumpFormat() {
	C.av_dump_format(ctx.CAVFormatContext, 0, nil, 0)
}

// AVFormatCloseInput closes an AVFormatContext input.
//
// The AVFormatContext parameter is a pointer to an AVFormatContext object.
// The function does not return any value.
func (ctx *AVFormatContext) AVFormatCloseInput() {
	C.avformat_close_input(&ctx.CAVFormatContext)
}

func (ctx *AVFormatContext) CheckIfNoFile(outStr string) {
	outCstr := C.CString(outStr)
	defer C.free(unsafe.Pointer(outCstr))

	//network connection do not need to open a file
	if (ctx.CAVFormatContext.flags & C.AVFMT_NOFILE) == 0 {
		ret := C.avio_open(&ctx.CAVFormatContext.pb, outCstr, C.AVIO_FLAG_WRITE)
		if ret < 0 {
			logger.Fatal("Output", zap.String("OutputOpenError", outStr))
		}
	}
}

func (ctx *AVFormatContext) AVFormatWriteHeader() int {
	return int(C.avformat_write_header(ctx.CAVFormatContext, nil))
}

// fucntion to call av_interleaved_write_frame
// CallAVInterleavedWriteFrame is a function to call av_interleaved_write_frame.
//
// Parameters:
//
//	ctx - The AVFormatContext object.
//	packet - The AVPacket object to be written.
//
// Returns:
//
//	int - The return value of av_interleaved_write_frame.
func (ctx *AVFormatContext) AVInterleavedWriteFrame(packet *AVPacket, stream *AVStream) int {
	if packet == nil || stream == nil {
		return 0
	}

	C.av_packet_rescale_ts(packet.CAVPacket, packet.CTimebase, stream.CAVStream.time_base)
	return int(C.av_interleaved_write_frame(ctx.CAVFormatContext, packet.CAVPacket))
}
