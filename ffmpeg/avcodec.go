package ffmpeg

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"

type AVCodecParameters struct {
	CAVCodecParameters *C.AVCodecParameters
}
