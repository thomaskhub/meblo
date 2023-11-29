package codecs

import (
	"errors"
	"fmt"
	"time"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
)

type Encoder struct {
	ctx            *astiav.CodecContext
	codec          *astiav.Codec
	dataChannel    chan astiav.Frame
	metaData       utils.MetaData
	packet         *astiav.Packet
	dataChannelOut chan astiav.Packet
	metaDataOut    utils.MetaData
	frameCnt       uint64
	startTime      time.Time
}

type EncoderStats struct {
	Fps   float64
	Speed float64
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) SetMetaData(metaData utils.MetaData) {
	e.metaData = metaData
}

func (e *Encoder) SetDataChannel(data chan astiav.Frame) {
	e.dataChannel = data
}

func (e *Encoder) GetDataChannel() chan astiav.Packet {
	return e.dataChannelOut
}

func (e *Encoder) GetMetaData() utils.MetaData {
	return e.metaDataOut
}

func (e *Encoder) GetStats() EncoderStats {
	fps := float64(e.frameCnt) / time.Since(e.startTime).Seconds()
	speed := fps / float64(e.metaData.FrameRate.Num())
	return EncoderStats{
		Fps:   fps,
		Speed: speed,
	}
}

func (e *Encoder) init() {
	e.dataChannelOut = make(chan astiav.Packet, 16)

	e.codec = astiav.FindEncoder(e.metaData.CodecPar.CodecID())
	if e.codec == nil {
		logger.Fatal("could not find codec")
	}

	e.ctx = astiav.AllocCodecContext(e.codec)
	if e.ctx == nil {
		logger.Fatal("could not alloc codec context")
	}

	e.packet = astiav.AllocPacket()
}

func (e *Encoder) open(opt *astiav.Dictionary) error {

	err := e.ctx.Open(e.codec, opt)
	if err != nil {
		logger.Fatal("could not open codec context")
	}

	codecPar := astiav.AllocCodecParameters()
	codecPar.FromCodecContext(e.ctx)

	e.metaDataOut = utils.MetaData{
		CodecPar: codecPar,
		TimeBase: e.ctx.TimeBase(),
	}

	return nil
}

func (e *Encoder) OpenVideo(bitrate int64, gopSize int, pixFormat astiav.PixelFormat, opt *astiav.Dictionary) error {
	e.init()

	e.ctx.SetHeight(e.metaData.CodecPar.Height())
	e.ctx.SetWidth(e.metaData.CodecPar.Width())
	e.ctx.SetSampleAspectRatio(e.metaData.CodecPar.SampleAspectRatio())
	e.ctx.SetPixelFormat(pixFormat)
	e.ctx.SetTimeBase(e.metaData.TimeBase)
	e.ctx.SetFramerate(e.metaData.FrameRate)

	e.ctx.SetBitRate(bitrate)
	e.ctx.SetGopSize(gopSize)

	e.ctx.SetFlags(e.ctx.Flags().Add(astiav.CodecContextFlagGlobalHeader))

	return e.open(opt)
}

func (e *Encoder) OpenAudio(
	bitrate int64, sampleFmt astiav.SampleFormat, chLayout astiav.ChannelLayout,
	opt *astiav.Dictionary) error {

	e.init()

	e.ctx.SetBitRate(bitrate)
	e.ctx.SetSampleFormat(sampleFmt)
	e.ctx.SetSampleRate(e.metaData.SampleRate)
	e.ctx.SetChannelLayout(chLayout)
	e.ctx.SetTimeBase(e.metaData.TimeBase)
	e.ctx.SetChannels(1)
	e.ctx.SetFlags(e.ctx.Flags().Add(astiav.CodecContextFlagGlobalHeader))

	return e.open(opt)
}

func (e *Encoder) Run() {
	e.startTime = time.Now()
	go func() {
		for {

			frame := <-e.dataChannel
			err := e.ctx.SendFrame(&frame)
			if err != nil {
				fmt.Printf("frame: %v\n", frame)
				logger.Fatal("could not send frame")
			}
			e.frameCnt++
			frame.Unref()

			err = e.ctx.ReceivePacket(e.packet)
			if err != nil {
				if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
					continue
				}
				logger.Fatal("could not receive packet")
			}

			e.dataChannelOut <- *e.packet.Clone()
			e.packet.Unref()
		}
	}()
}
