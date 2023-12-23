package chunker

import (
	"fmt"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/logger"
)

const (
	PTS_OFFSET_40_MS       = 4000
	PTS_OFFSET_AUDIO_FRAME = 2258
)

type FramePack struct {
	dataChannelOut chan astiav.Frame
	outCnt         uint64
	dataArr        []uint32
	ptsOffset      int64
	chunkSize      int
	sampleDiff     int
	removePts      bool
}

func NewFramePack(sampleDiff int) *FramePack {
	tmp := &FramePack{
		ptsOffset:      PTS_OFFSET_40_MS,
		chunkSize:      1764,
		dataChannelOut: make(chan astiav.Frame, 128),
		dataArr:        make([]uint32, 0),
		sampleDiff:     sampleDiff,
		removePts:      false,
	}

	if sampleDiff > 0 {
		//audio is greater then video so push empty audio frames

		adjust := make([]uint32, sampleDiff)
		for i := 0; i < sampleDiff; i++ {
			adjust[i] = 0
		}

		tmp.dataArr = append(tmp.dataArr, adjust...)

	} else if sampleDiff < 0 {
		//audio comes before video so cut of samples
		tmp.removePts = true
	}
	return tmp
}

func NewFrameUnPack(sampleDiff int) *FramePack {
	tmp := &FramePack{
		ptsOffset:      PTS_OFFSET_AUDIO_FRAME,
		chunkSize:      1024,
		dataChannelOut: make(chan astiav.Frame, 16),
		dataArr:        make([]uint32, 0),
		sampleDiff:     sampleDiff,
	}

	return tmp
}

func (f *FramePack) GetDataChannel() chan astiav.Frame {
	return f.dataChannelOut
}

func (f *FramePack) GetTimebase() astiav.Rational {
	return astiav.NewRational(1, 1000)
}

func (f *FramePack) PushFrame(frame *astiav.Frame) {

	if f.removePts {
		//this is not correct if the sampleDiff is very high this will actually fail
		logger.Fatal("TODO: need to implement when sampleDiff < 0")
		// f.removePts = false
		// d := frame.GetFrameData()
		// d = d[:f.sampleDiff] //cut pts values
		// f.dataArr = append(f.dataArr, d...)
	} else {
		f.dataArr = append(f.dataArr, frame.GetFrameData()...)
	}

	for len(f.dataArr) >= f.chunkSize {
		chunk := f.dataArr[:f.chunkSize]
		f.dataArr = f.dataArr[f.chunkSize:]

		outFrame := astiav.AllocFrame()
		outFrame.SetSampleFormat(astiav.SampleFormatFltp)
		outFrame.SetNbSamples(f.chunkSize)
		outFrame.SetChannelLayout(astiav.ChannelLayoutMono)
		outFrame.SetSampleRate(44100)
		outFrame.AllocBuffer(0)
		outFrame.SetFrameData(chunk)
		outFrame.SetPts(int64(f.outCnt) * f.ptsOffset)

		fmt.Printf("outFrame.Pts(): %v\n", outFrame.Pts())

		f.dataChannelOut <- *outFrame
		f.outCnt++
	}
}
