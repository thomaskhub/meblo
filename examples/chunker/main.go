package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/chunker"
	"github.com/thomaskhub/meblo/codecs"
	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap/zapcore"
)

func main() {
	//start profiler and create a file
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := cpuFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	logger.InitLogger(zapcore.ErrorLevel)
	astiav.SetLogLevel(astiav.LogLevelError)

	//setup the input
	in := inputs.NewInput()
	err = in.Open("rtmp://localhost:1935/live/test", false)
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

	//CHUNKER
	chunker := chunker.NewChunker()
	chunker.SetMetaData(videoDecoder.GetMetaData(), audioDecoder.GetMetaData())
	chunker.SetDataChannel(videoDecoder.GetDataChannel(), audioDecoder.GetDataChannel())

	//Start all the modules from back of the pipe to the beginning
	videoDecoder.Run()
	audioDecoder.Run()
	chunker.Run()
	in.Run()

	time.Sleep(time.Second * 30)
}
