package ffmpeg

//#cgo pkg-config: libavformat libavfilter libavutil libavcodec
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"

type AVCodecParameters struct {
	CAVCodecParameters *C.AVCodecParameters
}
