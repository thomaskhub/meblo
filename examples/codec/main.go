package main

import (
	"fmt"
	"time"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/decoders"
	"github.com/thomaskhub/meblo/encoders"
	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/outputs"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger.InitLogger(zapcore.ErrorLevel)
	// astiav.SetLogLevel(astiav.LogLevelDebug)

	//setup the input
	in := inputs.NewInput()
	err := in.Open("../../assets/test/test.mp4", false)
	if err != nil {
		panic(err)
	}

	//video decoding
	videoInMeta := in.GetMetaData()[inputs.OUT_VIDEO_CH]
	videoInData := (*in.GetDataChannel())[inputs.OUT_VIDEO_CH]
	videoDecoder := decoders.NewDecoder()
	videoDecoder.SetMetaData(videoInMeta)
	videoDecoder.SetDataChannel(videoInData)
	videoDecoder.Open()

	//encoding
	fmt.Printf("videoDecoder.GetMetaData(): %v\n", videoDecoder.GetMetaData().TimeBase)
	videoEncoder := encoders.NewEncoder()
	videoEncoder.SetMetaData(videoDecoder.GetMetaData())
	videoEncoder.SetDataChannel(videoDecoder.GetDataChannel())
	videoEncoder.Open(2500000, 50)

	//output
	out := outputs.NewOutput()
	outMap := make(map[int]chan astiav.Packet)
	outMap[0] = videoEncoder.GetDataChannel()

	outMeta := make(map[int]utils.MetaData)
	outMeta[0] = videoEncoder.GetMetaData()

	out.SetDataChannel(outMap)
	out.SetMetaData(outMeta)
	out.Open("/tmp/tomatoes.ts")

	//Start all the modules from back of the pipe to the beginning
	videoDecoder.Run()
	videoEncoder.Run()
	in.Run()

	// go func() {
	// 	for {
	// 		a := <-videoEncoder.GetDataChannel()
	// 		fmt.Printf("a.Duration(): %v\n", a.Duration())
	// 		// fmt.Printf("packet.Dts(): %v\n", packet.Dts())

	// 	}
	// }()

	go func() {
		//just ensure we are reading the audio packets from teh input so that system does nothange because of full fifo
		for {
			<-(*in.GetDataChannel())[inputs.OUT_AUDIO_CH]
			// println("Read audio packet")
		}

	}()

	time.Sleep(time.Second * 30)

}
