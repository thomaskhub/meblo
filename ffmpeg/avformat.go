package ffmpeg

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"
import (
	"log"
	"unsafe"

	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap"
)

// AVFormatContext represents an input or output format context.
type AVFormatContext struct {
	CAVFormatContext *C.AVFormatContext
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
func (ctx *AVFormatContext) GetStreams() []*AVStream {
	streams := make([]*AVStream, ctx.GetNumberOfStreams())
	numberOfStreams := ctx.GetNumberOfStreams()
	cStrams := (*C.AVStream)(unsafe.Pointer(ctx.CAVFormatContext.streams))

	for i := 0; i < numberOfStreams; i++ {

		arrayOfPointers := (*[1 << 30]*C.AVStream)(unsafe.Pointer(cStrams))[:numberOfStreams:numberOfStreams]
		stream := arrayOfPointers[i]
		streams[i] = &AVStream{CAVStream: stream}
	}

	return streams
}

// AVFormatAllocContext allocates an AVFormatContext struct.
//
// It returns a pointer to the allocated AVFormatContext struct.
func (ctx *AVFormatContext) AVFormatAllocContext() {
	ctx.CAVFormatContext = C.avformat_alloc_context()
}

func (ctx *AVFormatContext) AVFormatAllocOutputContext2(url string) int {
	//call the avformat_alloc_output_context2
	urlCsting := C.CString(url)
	defer C.free(unsafe.Pointer(urlCsting))

	// formatCstr := C.CString(utils.GetOutputFormat(url))
	// defer C.free(unsafe.Pointer(formatCstr))

	return int(C.avformat_alloc_output_context2(&ctx.CAVFormatContext, nil, nil, urlCsting))
}

func (ctx *AVFormatContext) AVFormatNewStream() *AVStream {
	return &AVStream{CAVStream: C.avformat_new_stream(ctx.CAVFormatContext, nil)}
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
	if ctx == nil || ctx.CAVFormatContext == nil {
		log.Println("-------------------- Pointer is nill")

	}
	return int(C.avformat_write_header(ctx.CAVFormatContext, nil))
}
