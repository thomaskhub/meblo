package outputs

import (
	"time"

	"github.com/thomaskhub/go-astiav"

	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

type Output struct {
	ctx         *astiav.FormatContext
	dataChannel utils.DataChannel
	metaData    utils.MetaDataChannel
	outStream   map[int]*astiav.Stream
	outStr      string
}

func NewOutput() *Output {
	return &Output{}
}

func (output *Output) SetMetaData(metaData utils.MetaDataChannel) {
	output.metaData = metaData
}

func (output *Output) SetDataChannel(dataChannel utils.DataChannel) {
	output.dataChannel = dataChannel
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
			for i, pktCh := range output.dataChannel {
				// fmt.Printf("i: %v\n", i)
				// packet := <-pktCh
				// fmt.Printf("after i: %v\n", i)
				// output.WriteInterleavedFrame(i, &packet)
				// packet.SetStreamIndex(i)
				select {

				case packet := <-pktCh:
					packet.SetStreamIndex(i)

					output.WriteInterleavedFrame(i, &packet)
					packet.Unref()

				default:

					time.Sleep(10 * time.Millisecond)
					// 	// No more audio packets currently available
				}
			}

		}
	}()

}

func (output *Output) Close() {
	output.ctx.CloseInput()
}
