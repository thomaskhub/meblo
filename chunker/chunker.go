package chunker

import (
	"fmt"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/utils"
)

type Chunker struct {
	videoDataChan chan astiav.Frame
	audioDataChan chan astiav.Frame
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

// func (f *Chunker) GetDataChannel() chan astiav.Frame {
// 	return f.dataChannelOut
// }

func (f *Chunker) GetMetaData() utils.MetaData {
	return f.metaDataOut
}

func NewChunker() *Chunker {
	return &Chunker{}
}

func (c *Chunker) Open() {

	c.videoDataChan = make(chan astiav.Frame, 16)
	c.audioDataChan = make(chan astiav.Frame, 16)
}

func (c *Chunker) Run() {
	isFirstPkt := true
	audioStarted := false
	videoStarted := false

	var frameCleanUp []*astiav.Frame
	frameCleanUp = make([]*astiav.Frame, 0)

	var curVFrame *astiav.Frame = nil
	var curAFrame *astiav.Frame = nil

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
					ptsDiff := curAFrame.Pts() - curVFrame.Pts()

					println("found av frame - align A-V")
					fmt.Printf("ptsDiff: %v\n", ptsDiff)

					//clean up the previous packets we have not processed
					for _, frame := range frameCleanUp {
						if frame != nil {
							frame.Unref()
						}
					}
					frameCleanUp = make([]*astiav.Frame, 0)

				} else {

				}
			} else {
				frameCleanUp = append(frameCleanUp, curVFrame)
				frameCleanUp = append(frameCleanUp, curAFrame)
			}

		}
	}()
}
