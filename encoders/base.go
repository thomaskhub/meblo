package encoders

import (
	"errors"
	"fmt"

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

func (e *Encoder) Open(bitrate int64, gopSize int) error {
	e.dataChannelOut = make(chan astiav.Packet, 16)

	e.codec = astiav.FindEncoder(e.metaData.CodecPar.CodecID())
	if e.codec == nil {
		logger.Fatal("could not find codec")
	}

	e.ctx = astiav.AllocCodecContext(e.codec)
	if e.ctx == nil {
		logger.Fatal("could not alloc codec context")
	}

	//TODO: this only works for video
	fmt.Printf("e.metaData.DecCtx.TimeBase(): %v\n", e.metaData.DecCtx.TimeBase())
	fmt.Printf("e.metaData.TimeBase: %v\n", e.metaData.TimeBase)
	// logger.Fatal("")

	e.ctx.SetHeight(e.metaData.CodecPar.Height())
	e.ctx.SetWidth(e.metaData.CodecPar.Width())
	e.ctx.SetSampleAspectRatio(e.metaData.CodecPar.SampleAspectRatio())
	e.ctx.SetPixelFormat(astiav.PixelFormatYuv420P) //TODO this is only valid for video,
	e.ctx.SetTimeBase(e.metaData.TimeBase)
	fmt.Printf("e.metaData.TimeBase: %v\n", e.metaData.TimeBase)
	// e.ctx.SetFramerate(astiav.NewRational(25, 1)) //TODO: this should come from global config
	// e.ctx.SetBitRate(bitrate)
	// e.ctx.SetGopSize(gopSize)
	// e.ctx.SetClorspace() --> needed?
	//color_trc ?
	//color_primaries?

	// e.ctx.SetFlags(e.ctx.Flags().Add(astiav.CodecContextFlagGlobalHeader))

	// dict := astiav.NewDictionary()
	// dict.Set()
	//TODO: we need to add support for av_opt_set to configure special encoder settings
	//   av_opt_set(this->ctx->priv_data, "level", "3.1", 0);
	//   av_opt_set(this->ctx->priv_data, "preset", "veryfast", 0);

	//   av_opt_set(this->ctx->priv_data, "profile", "high", 0);
	//   av_opt_set_int(this->ctx->priv_data, "sc_threshold", 0, 0);
	//   av_opt_set_int(this->ctx->priv_data, "forced-idr", 1, 0);
	e.packet = astiav.AllocPacket()

	err := e.ctx.Open(e.codec, nil)
	if err != nil {
		logger.Fatal("could not open codec context")
	}

	e.metaDataOut = utils.MetaData{
		CodecPar: e.metaData.CodecPar, //TODO: is this correct? Or must this be codec par from the output stream
		TimeBase: e.ctx.TimeBase(),
	}

	return nil
}

func (e *Encoder) Run() {
	go func() {
		for {

			frame := <-e.dataChannel
			err := e.ctx.SendFrame(&frame)
			if err != nil {
				fmt.Printf("frame: %v\n", frame)
				logger.Fatal("could not send frame")
			}

			//It seems that frame is ok, but packet seems to come twice sometim
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
