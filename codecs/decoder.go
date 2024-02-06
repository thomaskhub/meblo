package codecs

import (
	"errors"
	"sync"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
)

type Decoder struct {
	ctx            *astiav.CodecContext
	codec          *astiav.Codec
	decFrame       *astiav.Frame
	dataChannel    chan astiav.Packet
	metaData       utils.MetaData
	dataChannelOut chan astiav.Frame
	metaDataOut    utils.MetaData
	stop           chan bool
	wg             sync.WaitGroup
}

func NewDecoder() *Decoder {
	dec := &Decoder{
		stop: make(chan bool),
	}
	dec.wg.Add(1)
	return dec
}

func (e *Decoder) SetMetaData(metaData utils.MetaData) {
	e.metaData = metaData
}

func (e *Decoder) SetDataChannel(data chan astiav.Packet) {
	e.dataChannel = data
}

func (e *Decoder) GetDataChannel() chan astiav.Frame {
	return e.dataChannelOut
}

func (e *Decoder) GetMetaData() utils.MetaData {
	return e.metaDataOut
}

func (e *Decoder) Free() {
	e.stop <- true
	e.wg.Wait()
	e.ctx.Free()
}

func (e *Decoder) Open() {

	e.codec = astiav.FindDecoder(e.metaData.CodecPar.CodecID())
	if e.codec == nil {
		logger.Fatal("could not find codec")
	}

	e.ctx = astiav.AllocCodecContext(e.codec)
	if e.ctx == nil {
		logger.Fatal("could not alloc codec context")
	}

	e.metaData.CodecPar.ToCodecContext(e.ctx)

	err := e.ctx.Open(e.codec, nil)
	if err != nil {
		logger.Fatal("could not open codec context")
	}

	e.ctx.SetFramerate(e.metaData.FrameRate)
	e.ctx.SetTimeBase(e.metaData.TimeBase)

	e.decFrame = astiav.AllocFrame()
	e.dataChannelOut = make(chan astiav.Frame, 16)

	e.metaDataOut.CodecPar = e.metaData.CodecPar
	e.metaDataOut.TimeBase = e.metaData.TimeBase
	e.metaDataOut.FrameRate = e.metaData.FrameRate
	e.metaDataOut.SampleRate = e.metaData.SampleRate

}

func (e *Decoder) Run() {
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-e.stop:
				return
			default:

				packet := <-e.dataChannel
				err := e.ctx.SendPacket(&packet)
				if err != nil {
					logger.Fatal("could not send packet")
				}
				packet.Unref()
				packet.Free()

				err = e.ctx.ReceiveFrame(e.decFrame)
				if err != nil {
					if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
						continue
					}
					logger.Fatal("could not receive frame")
				}

				//when we com here we can push the frame to the output
				e.dataChannelOut <- *e.decFrame.Clone() //TODO: having clone here make things work !!!!
				e.decFrame.Unref()
			}
		}
	}()
}
