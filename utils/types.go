package utils

import "github.com/thomaskhub/go-astiav"

const (
	OutputModeRtmp   = "flv"
	OutputModeMpegts = "mpegts"
	OutputModeDash   = "dash"
)

type MetaData struct {
	TimeBase   astiav.Rational
	CodecPar   *astiav.CodecParameters
	FrameRate  astiav.Rational
	SampleRate int
}

type Chunk struct {
	VideoFrame    astiav.Frame
	AudioFrame    astiav.Frame
	AudioCodecPar *astiav.CodecParameters
	VideoCodecPar *astiav.CodecParameters
}

type MetaDataChannel map[int]MetaData

type DataChannel map[int]chan astiav.Packet

// type DataChannel map[int]chan *astiav.Packet
