package chunker

import (
	"github.com/thomaskhub/go-astiav"
)

const (
	PTS_OFFSET_40_MS       = 40000
	PTS_OFFSET_AUDIO_FRAME = 22575
)

type FramePack struct {
	dataChannelOut chan astiav.Frame
	outCnt         uint64
	dataArr        []uint32
	ptsOffset      int64
	chunkSize      int
	ptsDiff        int
	removePts      bool
}

func NewFramePack(ptsDiff int, frame *astiav.Frame) *FramePack {
	tmp := &FramePack{
		ptsOffset:      PTS_OFFSET_40_MS,
		chunkSize:      1764,
		dataChannelOut: make(chan astiav.Frame, 16),
		dataArr:        make([]uint32, 0),
		ptsDiff:        ptsDiff,
		removePts:      false,
	}

	if ptsDiff > 0 {
		//audio is greater then video so push empty audio frames

		adjust := make([]uint32, 0)
		for i := 0; i < ptsDiff; i++ {
			adjust[i] = 0
		}

		tmp.dataArr = append(tmp.dataArr, adjust...)

	} else if ptsDiff < 0 {
		//audio comes before video so cut of samples
		tmp.removePts = true
	}
	return tmp
}

func NewFrameUnPack(ptsDiff int, frame *astiav.Frame) *FramePack {
	tmp := &FramePack{
		ptsOffset:      PTS_OFFSET_AUDIO_FRAME,
		chunkSize:      1024,
		dataChannelOut: make(chan astiav.Frame, 16),
		dataArr:        make([]uint32, 0),
		ptsDiff:        ptsDiff,
	}

	return tmp
}

func (f *FramePack) GetDataChannel() chan astiav.Frame {
	return f.dataChannelOut
}

func (f *FramePack) PushFrame(frame *astiav.Frame) {

	if f.removePts {
		f.removePts = false
		d := frame.GetFrameData()
		d = d[:f.ptsDiff] //cut pts values
		f.dataArr = append(f.dataArr, d...)
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

		f.dataChannelOut <- *outFrame
		f.outCnt++
	}
}
