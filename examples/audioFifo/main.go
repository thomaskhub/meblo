package main

import (
	"fmt"
	"os"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/codecs"
	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap/zapcore"
)

func main() {

	logger.InitLogger(zapcore.ErrorLevel)
	// astiav.SetLogLevel(astiav.LogLevelDebug)

	in := inputs.NewInput()
	err := in.Open("../../assets/test/test.mp4", false)
	if err != nil {
		panic(err)
	}

	//DECODERS
	audioInMeta := in.GetMetaData()[inputs.OUT_AUDIO_CH]
	audioInData := (*in.GetDataChannel())[inputs.OUT_AUDIO_CH]
	audioDecoder := codecs.NewDecoder()
	audioDecoder.SetMetaData(audioInMeta)
	audioDecoder.SetDataChannel(audioInData)
	audioDecoder.Open()

	videoInMeta := in.GetMetaData()[inputs.OUT_VIDEO_CH]
	videoInData := (*in.GetDataChannel())[inputs.OUT_VIDEO_CH]
	videoDecoder := codecs.NewDecoder()
	videoDecoder.SetMetaData(videoInMeta)
	videoDecoder.SetDataChannel(videoInData)
	videoDecoder.Open()

	//Start all the modules from back of the pipe to the beginning
	videoDecoder.Run()
	audioDecoder.Run()
	in.Run()

	//Audio fifo
	fifo := astiav.AllocateAudioFifo(astiav.SampleFormatFlt, 2, 2*1764)
	aCh := audioDecoder.GetDataChannel()

	for i := 0; i < 5; i++ {
		frame := <-aCh
		fmt.Printf("frame.NbSamples(): %v\n", frame.NbSamples())
		err := fifo.PushFrame(&frame)
		if err != nil {
			fmt.Println("Error", err)
			os.Exit(1)
		}
	}

	fmt.Printf("fifo.Size(): %v\n", fifo.Size())

	// for {
	// print("try to pull frame\n")
	// frame := astiav.AllocFrame()

	frame := astiav.AllocFrame()
	for i := 0; i < 3; i++ {
		if fifo.Size() >= 1764 {
			frame.SetNbSamples(1764)
			frame.SetSampleFormat(astiav.SampleFormatFlt)
			frame.SetSampleRate(44100)
			frame.SetChannelLayout(astiav.ChannelLayoutStereo)

			ret := fifo.PullFrame(frame)
			fmt.Printf("ret: %v\n", ret)
			// if err != nil {
			// 	fmt.Println("Error", err)
			// 	// break
			// }
			// frame.Free()
			// fmt.Printf("fifo.Size(): %v\n", fifo.Size())
		}
	}

	// fmt.Printf("err: %v\n", err)
	// fmt.Printf("frame.NbSamples(): %v\n", frame.NbSamples())

	// if err != nil {
	// 	fmt.Println("Error", err)
	// 	// break
	// }
	// // }

	// time.Sleep(time.Second * 30)
	in.Close()
	audioDecoder.Free()
	videoDecoder.Free()
	fifo.Free()
}
