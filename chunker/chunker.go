package chunker

import (
	"fmt"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/utils"
)

type Chunker struct {
	videoDataChan chan astiav.Frame
	audioDataChan chan astiav.Frame
	chunkDataOut  chan utils.Chunk
	videoMetaData utils.MetaData
	audioMetaData utils.MetaData
	metaDataOut   utils.MetaData
}

func (c *Chunker) SetMetaData(videoMetaData utils.MetaData, audioMetaData utils.MetaData) {
	c.videoMetaData = videoMetaData
	c.audioMetaData = audioMetaData
}

func (c *Chunker) SetDataChannel(video chan astiav.Frame, audio chan astiav.Frame) {
	c.videoDataChan = video
	c.audioDataChan = audio
}

func (f *Chunker) GetDataChannel() chan utils.Chunk {
	return f.chunkDataOut
}

func (f *Chunker) GetMetaData() utils.MetaData {
	return f.metaDataOut
}

func NewChunker() *Chunker {
	return &Chunker{}
}

func (c *Chunker) Open() {

	c.chunkDataOut = make(chan utils.Chunk, 16)

	c.metaDataOut = utils.MetaData{
		TimeBase: astiav.NewRational(1, 1000),
	}
	println("open chunker")
	// c.videoDataChan = make(chan astiav.Frame, 16)
	// c.audioDataChan = make(chan astiav.Frame, 16)
	// c.chunkDataOut = make(chan utils.Chunk, 16)

	// println("open chunker 12")
	// c.metaDataOut = utils.MetaData{
	// 	TimeBase: astiav.NewRational(1, 1000),
	// }
	println("open chunker312")
}

// only works with 25fps!!!
func (c *Chunker) Run() {
	isFirstPkt := true
	audioStarted := false
	videoStarted := false
	videoOutCnt := 0

	var frameCleanUp []*astiav.Frame
	frameCleanUp = make([]*astiav.Frame, 0)

	var curVFrame *astiav.Frame = nil
	var curAFrame *astiav.Frame = nil

	interVideoCh := make(chan *astiav.Frame, 16)

	var pack *FramePack = nil

	go func() {
		for {

			select {
			case audioFrame := <-c.audioDataChan:
				curAFrame = &audioFrame
				audioStarted = true
			default:

				//No audio or video frame found
			}

			select {
			case videoFrame := <-c.videoDataChan:
				curVFrame = &videoFrame
				videoStarted = true
			default:

				//No audio or video frame found
			}

			if audioStarted && videoStarted {
				//Now we can start processing as we have video and audio packets available

				if isFirstPkt {
					isFirstPkt = false

					curAFrame.SetPts(astiav.RescaleQ(curAFrame.Pts(), c.audioMetaData.TimeBase, c.videoMetaData.TimeBase))

					// aPtsInMs := float64(curAFrame.Pts()) * float64(c.audioMetaData.TimeBase.Num()) / float64(c.audioMetaData.TimeBase.Den())
					// vPtsInMs := float64(curVFrame.Pts()) * float64(c.videoMetaData.TimeBase.Num()) / float64(c.videoMetaData.TimeBase.Den())
					ptsDiff := curAFrame.Pts() - curVFrame.Pts()

					fmt.Printf("ptsDiff: %v\n", ptsDiff)

					sampleTime := float64(c.audioMetaData.TimeBase.Den()) / float64(c.audioMetaData.SampleRate)
					sampleDiff := int(float64(ptsDiff) / sampleTime)

					pack = NewFramePack(sampleDiff)

					println("Going to push frame ")
					pack.PushFrame(curAFrame)
					println("Going to pus frame in video chanel ")
					interVideoCh <- curVFrame

					println("found av frame - align A-V")
					// fmt.Printf("ptsDiff: %v\n", ptsDiff)

					//clean up the previous packets we have not processed
					for _, frame := range frameCleanUp {
						if frame != nil {
							frame.Unref()
						}
					}
					frameCleanUp = make([]*astiav.Frame, 0)

				} else {
					if curVFrame != nil {
						interVideoCh <- curVFrame
					}

					if curAFrame != nil {
						pack.PushFrame(curAFrame)
					}
				}
			} else {
				frameCleanUp = append(frameCleanUp, curVFrame)
				frameCleanUp = append(frameCleanUp, curAFrame)
			}

		}
	}()

	go func() {
		for {

			videoFrame := <-interVideoCh
			audioFrame := <-pack.GetDataChannel()

			videoFrame.SetPts(int64(videoOutCnt) * PTS_OFFSET_40_MS)
			videoOutCnt++
			// fmt.Printf("videoFrame.Pts(): %v\n", videoFrame.Pts())
			// fmt.Printf("c.videoMetaData.TimeBase: %v\n", c.videoMetaData.TimeBase)
			// fmt.Printf("pack.GetTimebase(): %v\n", pack.GetTimebase())

			// rescVal := astiav.RescaleQ(videoFrame.Pts(), c.videoMetaData.TimeBase, pack.GetTimebase())
			// fmt.Printf("rescVal: %v\n", rescVal)
			// videoFrame.SetPts(rescVal)

			//now audio and video chunks are aligned and 40ms in length so we can
			//give them to the next state

			c.chunkDataOut <- utils.Chunk{
				VideoFrame:    *videoFrame.Clone(),
				AudioFrame:    *audioFrame.Clone(),
				AudioCodecPar: c.audioMetaData.CodecPar,
				VideoCodecPar: c.videoMetaData.CodecPar,
			}

			// videoFrame.Unref()
			// audioFrame.Unref()
		}
	}()
}
