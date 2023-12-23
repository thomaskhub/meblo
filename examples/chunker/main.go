package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/chunker"
	"github.com/thomaskhub/meblo/codecs"
	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/outputs"
	"github.com/thomaskhub/meblo/utils"
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
	astiav.SetLogLevel(astiav.LogLevelDebug)

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
	chuk := chunker.NewChunker()
	chuk.SetMetaData(videoDecoder.GetMetaData(), audioDecoder.GetMetaData())
	chuk.SetDataChannel(videoDecoder.GetDataChannel(), audioDecoder.GetDataChannel())
	chuk.Open()

	chunkStream := chuk.GetDataChannel()
	encVData := make(chan astiav.Frame, 16)
	encAData := make(chan astiav.Frame, 16)

	unpack := chunker.NewFrameUnPack(0)
	go func() {
		//read the chunk stream and push it to audio video encoder
		for {

			chunk := <-chunkStream
			fmt.Printf("\"read chunk\": %v\n", "read chunk")
			encVData <- chunk.VideoFrame

			unpack.PushFrame(&chunk.AudioFrame)
			d := <-unpack.GetDataChannel()
			encAData <- d
		}
	}()

	//ENCODER
	audioEncoder := codecs.NewEncoder()
	audioEncoder.SetMetaData(audioDecoder.GetMetaData())

	// audioEncoder.SetDataChannel(audioDecoder.GetDataChannel())
	audioEncoder.SetDataChannel(encAData)
	audioEncoder.OpenAudio(64000, astiav.SampleFormatFltp, astiav.ChannelLayoutMono, nil)

	vidoeEncOpts := astiav.NewDictionary()
	vidoeEncOpts.Set("level", "3.1", 0)
	vidoeEncOpts.Set("preset", "veryfast", 0)
	vidoeEncOpts.Set("crf", "23", 0)
	vidoeEncOpts.Set("profile", "high", 0)
	vidoeEncOpts.Set("sc_threshold", "0", 0)
	vidoeEncOpts.Set("forced-idr", "1", 0)
	videoEncoder := codecs.NewEncoder()
	videoEncoder.SetMetaData(videoDecoder.GetMetaData())
	// videoEncoder.SetDataChannel(videoDecoder.GetDataChannel())
	videoEncoder.SetDataChannel(encVData)
	videoEncoder.OpenVideo(2500000, 50, astiav.PixelFormatYuv420P, vidoeEncOpts)

	//output
	out := outputs.NewOutput()
	outMap := make(map[int]chan astiav.Packet)
	outMap[0] = videoEncoder.GetDataChannel()
	outMap[1] = audioEncoder.GetDataChannel()

	outMeta := make(map[int]utils.MetaData)
	outMeta[0] = videoEncoder.GetMetaData()
	outMeta[1] = audioEncoder.GetMetaData()

	out.SetDataChannel(outMap)
	out.SetMetaData(outMeta)
	out.Open("/tmp/cucumber.ts")

	//Start all the modules from back of the pipe to the beginning
	videoDecoder.Run()
	audioDecoder.Run()
	chuk.Run()
	audioEncoder.Run()
	videoEncoder.Run()
	in.Run()

	time.Sleep(time.Second * 30)
}
