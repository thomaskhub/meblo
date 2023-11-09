package ffmpeg

//#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu
//#cgo LDFLAGS: -lavformat -lavcodec -lavutil
//#include <libavformat/avformat.h>
// #include <errno.h>
import "C"

// AVPacket represents a packet of audio or video data.
type AVPacket struct {
	CAVPacket *C.AVPacket
}

func (packet *AVPacket) AVPacketAllocate() {
	packet.CAVPacket = C.av_packet_alloc()
}

func (packet *AVPacket) GetDuration() int {
	return int(packet.CAVPacket.duration)
}

func (packet *AVPacket) GetStreamIndex() int {
	return int(packet.CAVPacket.stream_index)
}

func (packet *AVPacket) AVPacketFree() {
	if packet == nil || packet.CAVPacket == nil {
		return
	}
	C.av_packet_free(&packet.CAVPacket)
}
