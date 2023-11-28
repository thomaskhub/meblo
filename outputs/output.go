package outputs

import (
	"github.com/thomaskhub/go-astiav"

	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type Output struct {
	ctx         *astiav.FormatContext
	dataChannel *utils.DataChannel
	outStream   map[int]*astiav.Stream
	metaData    utils.MetaDataChannel
	outStr      string
}

func NewOutput() *Output {
	return &Output{}
}

func (output *Output) SetMetaData(metaData utils.MetaDataChannel) {
	output.metaData = metaData
}

func (output *Output) SetDataChannel(dataChannel utils.DataChannel) {
	output.dataChannel = &dataChannel
}

func (output *Output) CheckIfNoFile() *astiav.IOContext {
	noFile := output.ctx.OutputFormat().Flags().Has(astiav.IOFormatFlagNofile)
	if !noFile {
		ioContext := astiav.NewIOContext()
		flags := astiav.NewIOContextFlags(astiav.IOContextFlagWrite)
		if err := ioContext.Open(output.outStr, flags); err != nil {
			logger.Fatal("Output", zap.String("OutputOpenError", output.outStr))
		}
		output.ctx.SetPb(ioContext)
		return ioContext
	}

	return nil
}

func (output *Output) WriteInterleavedFrame(idx int, packet *astiav.Packet) {

	outTimeBase := output.outStream[idx].TimeBase()
	inTimeBase := output.metaData[idx].TimeBase

	packet.RescaleTs(inTimeBase, outTimeBase)
	// packet.SetPos(-1) --> this we never used in the past is this needed?

	if err := output.ctx.WriteInterleavedFrame(packet); err != nil {
		logger.Fatal("Output", zap.String("OutputWriteError", output.outStr))
	}

}

func (output *Output) Open(outStr string) {
	output.outStr = outStr
	var err error

	output.outStream = make(map[int]*astiav.Stream)

	output.ctx, err = astiav.AllocOutputFormatContext(nil, "mpegts", outStr)
	if err != nil {
		logger.Fatal("Output", zap.String("OutputFormatError", outStr))
	}

	//create the streams for video
	for i, metaData := range output.metaData {
		os := output.ctx.NewStream(nil)
		if os == nil {
			logger.Fatal("Output", zap.String("OutputFormatError", outStr))
		}

		output.outStream[i] = os
		metaData.CodecPar.Copy(os.CodecParameters())
	}

	output.CheckIfNoFile()

	output.ctx.WriteHeader(nil)

	//It might be that write interleave write frame has some concurrency issues
	//so thats why output audio and video processing is done in one go routine
	go func() {
		for {
			for i, pktCh := range *output.dataChannel {
				select {

				case packet := <-pktCh:
					output.WriteInterleavedFrame(i, packet)
					packet.Unref()

				default:
					// No more audio packets currently available
				}
			}

		}
	}()

}

func (output *Output) Close() {
	output.ctx.CloseInput()
}
