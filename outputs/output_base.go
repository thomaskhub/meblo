package outputs

import (
	"fmt"

	"github.com/thomaskhub/meblo/ffmpeg"
	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap"
)

const (
	OutputModeVideo = "video"
	OutputModeAudio = "audio"
)

type OutputBase struct {
	ctx                   *ffmpeg.AVFormatContext
	videoCodecParamsList  []ffmpeg.AVCodecParameters
	audioCodecParamsList  []ffmpeg.AVCodecParameters
	videoOutputStreamList []*ffmpeg.AVStream
	audioOutputStreamList []*ffmpeg.AVStream
	videoIn               [](chan *ffmpeg.AVPacket)
	audioIn               [](chan *ffmpeg.AVPacket)
	outStr                string
}

func (output *OutputBase) ConfigVideo(out chan *ffmpeg.AVPacket, codecParameters ffmpeg.AVCodecParameters) {
	output.videoIn = append(output.videoIn, out)
	output.videoCodecParamsList = append(output.videoCodecParamsList, codecParameters)
}

func (output *OutputBase) ConfigAudio(out chan *ffmpeg.AVPacket, codecParameters ffmpeg.AVCodecParameters) {
	output.audioIn = append(output.audioIn, out)
	output.audioCodecParamsList = append(output.audioCodecParamsList, codecParameters)
}

func (output *OutputBase) Open(outStr string) {
	output.outStr = outStr
	output.ctx = &ffmpeg.AVFormatContext{}

	//open ffmpeg output context 2
	output.ctx.AVFormatAllocContext()
	ret := output.ctx.AVFormatAllocOutputContext2(outStr)
	if ret < 0 {
		logger.Fatal("Output", zap.String("OutputFormatError", outStr))
	}

	//create the streams for video

	fmt.Printf("output.videoCodecParamsList: %v\n", output.videoCodecParamsList[0].CAVCodecParameters)

	for i, codecParams := range output.videoCodecParamsList {
		output.videoOutputStreamList = append(output.videoOutputStreamList, output.ctx.AVFormatNewStream())
		output.videoOutputStreamList[i].SetCodecParameters(codecParams)

		fmt.Printf("codecParams: %v\n", codecParams)

	}

	for i, codecParams := range output.audioCodecParamsList {
		output.audioOutputStreamList = append(output.audioOutputStreamList, output.ctx.AVFormatNewStream())
		output.audioOutputStreamList[i].SetCodecParameters(codecParams)

	}

	output.ctx.CheckIfNoFile(outStr)
	output.ctx.AVFormatWriteHeader()

	//process audio packets
	go func() {
		for {
			for i, out := range output.audioIn {
				packet := <-out
				fmt.Printf("packet: %v\n", packet)
				stream := output.audioOutputStreamList[i]

				// packet.PtsSwitchDts()
				ret := output.ctx.AVInterleavedWriteFrame(packet, stream)
				if ret < 0 {
					logger.Fatal("Output Audio", zap.String("OutputWriteError", outStr))
				}

				packet.AVPacketUnref()
			}
		}
	}()

	// //process vide packets
	go func() {
		for {
			for i, out := range output.videoIn {
				packet := <-out

				// fmt.Printf("packet.CAVPacket: %v %d\n", packet.CAVPacket, i)
				stream := output.videoOutputStreamList[i]
				// packet.PtsSwitchDts()

				output.ctx.AVInterleavedWriteFrame(packet, stream)
				if ret < 0 {
					logger.Fatal("Output Video", zap.String("OutputWriteError", outStr))
				}

				packet.AVPacketUnref()
			}
		}
	}()

}

func (output *OutputBase) Close() {
	output.ctx.AVFormatFreeContext()
}
