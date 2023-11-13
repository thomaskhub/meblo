package ffmpeg

//#cgo pkg-config: libavformat libavfilter libavutil libavcodec
//#include <libavformat/avformat.h>
//#include <errno.h>
import "C"
import "unsafe"

// AVPacket represents a packet of audio or video data.
type AVPacket struct {
	CAVPacket *C.AVPacket
	CTimebase C.AVRational
}

// clone the packet. This is helpful if you need to send the same packets to multiple outputs
func (packet *AVPacket) AVPacketClone() *AVPacket {
	return &AVPacket{
		CAVPacket: C.av_packet_clone(packet.CAVPacket),
		CTimebase: packet.CTimebase,
	}
}

// switch the pts and dts timestamp of a packet
func (packet *AVPacket) PtsSwitchDts() {
	tmp := packet.CAVPacket.pts
	packet.CAVPacket.pts = C.int64_t(packet.CAVPacket.dts)
	packet.CAVPacket.dts = C.int64_t(tmp)
}

// AVPacketAllocate allocates an AVPacket.
func (packet *AVPacket) AVPacketAllocate() {
	unsafePointer := unsafe.Pointer(C.av_packet_alloc())
	packet.CAVPacket = (*C.AVPacket)(unsafePointer)
}

// GetDuration returns the duration of the packet.
func (packet *AVPacket) GetDuration() int {
	return int(packet.CAVPacket.duration)
}

func (packet *AVPacket) GetPts() int {
	return int(packet.CAVPacket.pts)
}

func (packet *AVPacket) GetDts() int {
	return int(packet.CAVPacket.dts)
}

func (stream *AVPacket) SetTimebase(t Timescale) {
	stream.CTimebase = C.AVRational{
		num: C.int(t.Num),
		den: C.int(t.Den),
	}
}

// GetStreamIndex returns the stream index of the packet.
func (packet *AVPacket) GetStreamIndex() int {
	return int(packet.CAVPacket.stream_index)
}

// AVPacketFree frees an AVPacket.
func (packet *AVPacket) AVPacketFree() {
	if packet == nil || packet.CAVPacket == nil {
		return
	}
	C.av_packet_free(&packet.CAVPacket)
}

// Unref a packet so that it can be freed by libav if not needed any longer, needs to be called only in the output module
func (packet *AVPacket) AVPacketUnref() {
	C.av_packet_unref(packet.CAVPacket)
}
