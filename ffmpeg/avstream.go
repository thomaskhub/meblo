package ffmpeg

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"

type AVStream struct {
	CAVStream *C.AVStream
}

func (stream *AVStream) GetCodecPar() AVCodecParameters {
	return AVCodecParameters{CAVCodecParameters: stream.CAVStream.codecpar}
}

func (stream *AVStream) GetStreamIndex() int {
	return int(stream.CAVStream.index)
}

func (stream *AVStream) GetCodecType() int {
	return int(stream.CAVStream.codecpar.codec_type)
}

func (stream *AVStream) GetTimebase() Timescale {
	return Timescale{Num: int(stream.CAVStream.time_base.num), Den: int(stream.CAVStream.time_base.den)}
}

func (stream *AVStream) SetCodecParameters(codecpar AVCodecParameters) {
	stream.CAVStream.codecpar = codecpar.CAVCodecParameters
}
