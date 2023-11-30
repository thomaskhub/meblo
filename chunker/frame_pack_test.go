package chunker_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/chunker"
	"github.com/thomaskhub/meblo/logger"
	"go.uber.org/zap/zapcore"
)

func PrintData(frame *astiav.Frame) {
	for _, val := range frame.GetFrameData() {
		fmt.Printf("%08d,", val)
	}
}

func SetFrameData(t *testing.T, frame *astiav.Frame, startVal int) {

	frame.SetMonoTestData(startVal)

	//test that data has been set correctly
	for i, val := range frame.GetFrameData() {
		require.Equal(t, uint32(i+startVal), val)
	}

}

func TestFramePack(t *testing.T) {
	logger.InitLogger(zapcore.DebugLevel)

	f1 := astiav.AllocFrame()
	f1.SetSampleFormat(astiav.SampleFormatFltp)
	f1.SetNbSamples(1024)
	f1.SetChannelLayout(astiav.ChannelLayoutMono)
	f1.SetSampleRate(44100)
	f1.AllocBuffer(0)
	f1.SetPts(1234)

	pack := chunker.NewFramePack(0, nil)
	for i := 0; i < 4; i++ {
		SetFrameData(t, f1, 1024*i)
		pack.PushFrame(f1)
	}

	done := false
	k := 0
	chunkList := make([]astiav.Frame, 0)
	for !done {
		select {
		case chunk, ok := <-pack.GetDataChannel():
			if !ok {
				done = true
			}

			d := chunk.GetFrameData()

			require.Equal(t, len(d), 1764)
			for i, val := range d {
				require.Equal(t, uint32(i+k*1764), val)
			}

			require.Equal(t, chunk.Pts(), int64(0+k*40000))
			chunkList = append(chunkList, chunk)
		default:
			done = true
		}

		k++
	}

	unpack := chunker.NewFrameUnPack(0, nil)

	for _, chunk := range chunkList {
		unpack.PushFrame(&chunk)
	}

	done = false
	k = 0
	for !done {
		select {
		case chunk, ok := <-unpack.GetDataChannel():
			if !ok {
				done = true
			}

			d := chunk.GetFrameData()

			require.Equal(t, len(d), 1024)
			for i, val := range d {
				require.Equal(t, uint32(i+k*1024), val)
			}

			require.Equal(t, chunk.Pts(), int64(0+k*chunker.PTS_OFFSET_AUDIO_FRAME))

		default:
			done = true
		}

		k++
	}
}

func TestFramePackPosPts(t *testing.T) {
	logger.InitLogger(zapcore.DebugLevel)
	logger.Debug("TestFramePackPosPts")
}

func TestFramePackNegPts(t *testing.T) {
	logger.InitLogger(zapcore.DebugLevel)
	logger.Debug("TestFramePackNegPts")
}
