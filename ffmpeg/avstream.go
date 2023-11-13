package ffmpeg

//#cgo pkg-config: libavformat libavfilter libavutil libavcodec
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"
import (
	"fmt"
	"unsafe"
)

type AVStream struct {
	CAVStream *C.AVStream
}

func (stream *AVStream) GetCodecPar() AVCodecParameters {
	cCodecPar := unsafe.Pointer(stream.CAVStream.codecpar)
	return AVCodecParameters{CAVCodecParameters: (*C.AVCodecParameters)(cCodecPar)}
}

func (stream *AVStream) GetStreamIndex() int {
	return int(stream.CAVStream.index)
}

func (stream *AVStream) GetCodecType() int {
	fmt.Printf("stream.CAVStream.codecpar: %v\n", stream.CAVStream.codecpar.codec_type)
	return int(stream.CAVStream.codecpar.codec_type)
}

func (stream *AVStream) GetTimebase() Timescale {
	return Timescale{Num: int(stream.CAVStream.time_base.num), Den: int(stream.CAVStream.time_base.den)}
}

func (stream *AVStream) SetCodecParameters(codecpar AVCodecParameters) {
	C.avcodec_parameters_copy(stream.CAVStream.codecpar, codecpar.CAVCodecParameters)
	// fmt.Printf("stream.CAVStream.codecpar: %v\n", stream.CAVStream.codec.codec_tag)
}
