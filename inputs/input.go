package inputs

import (
	"errors"
	"fmt"
	"sync"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
	"go.uber.org/zap"
)

const (
	OUT_AUDIO_CH = 0
	OUT_VIDEO_CH = 1
)

type Input struct {
	ctx *astiav.FormatContext

	//Channels for the output
	dataChannel utils.DataChannel

	//Streams from the input file
	videoStreams []*astiav.Stream
	audioStreams []*astiav.Stream

	//side band signals needed for the output if no encoder is used
	metaData utils.MetaDataChannel
	stop     chan bool
	wg       sync.WaitGroup
}

func NewInput() *Input {
	in := &Input{}
	in.wg.Add(1)
	in.stop = make(chan bool)
	return in
}

func (in *Input) Open(url string, autoRetry bool) error {

	in.ctx = astiav.AllocFormatContext()
	// err := in.ctx.OpenInput(url, astiav.FindInputFormat("flv"), nil)
	err := in.ctx.OpenInput(url, nil, nil)
	if err != nil {
		logger.Fatal("could not open input")
	}

	err = in.ctx.FindStreamInfo(nil)
	if err != nil {
		logger.Fatal("could not find stream info")
	}

	for _, stream := range in.ctx.Streams() {
		fmt.Printf("stream: %v\n", stream)
		if stream.CodecParameters().CodecType() == astiav.MediaTypeVideo {
			in.videoStreams = append(in.videoStreams, stream)
		} else if stream.CodecParameters().CodecType() == astiav.MediaTypeAudio {
			in.audioStreams = append(in.audioStreams, stream)
		}
	}

	err = in.CheckHasStreams(1, 1)
	if err != nil {
		logger.Error("FileInput", zap.String("StreamCheck", "failed"))
		return err
	}

	in.dataChannel = make(utils.DataChannel)
	in.dataChannel[OUT_VIDEO_CH] = make(chan astiav.Packet, 16)
	in.dataChannel[OUT_AUDIO_CH] = make(chan astiav.Packet, 16)

	in.metaData = make(utils.MetaDataChannel)

	in.metaData[OUT_VIDEO_CH] = utils.MetaData{
		TimeBase:  in.videoStreams[0].TimeBase(),
		CodecPar:  in.videoStreams[0].CodecParameters(),
		FrameRate: in.ctx.GuessFrameRate(in.videoStreams[0], nil),
	}

	in.metaData[OUT_AUDIO_CH] = utils.MetaData{
		TimeBase:   in.audioStreams[0].TimeBase(),
		CodecPar:   in.audioStreams[0].CodecParameters(),
		SampleRate: in.audioStreams[0].CodecParameters().SampleRate(),
	}

	return nil
}

func (in *Input) CheckHasStreams(numberVideoStreams, numberAudioStreams int) error {
	if len(in.videoStreams) < int(numberVideoStreams) || len(in.audioStreams) < int(numberAudioStreams) {
		return fmt.Errorf("number of video or audio streams does not match the specified count")
	}
	return nil
}

func (in *Input) Run() {
	go func() {
		defer in.wg.Done()
		for {
			select {
			case <-in.stop:
				return

			default:

				packet := astiav.AllocPacket()
				err := in.ctx.ReadFrame(packet)
				if err != nil {
					if errors.Is(err, astiav.ErrEof) {
						break
					}
					logger.Fatal("could not read frame")
				}

				if packet.StreamIndex() == in.videoStreams[0].Index() {
					in.dataChannel[OUT_VIDEO_CH] <- *packet.Clone()
				} else if packet.StreamIndex() == in.audioStreams[0].Index() {
					in.dataChannel[OUT_AUDIO_CH] <- *packet.Clone()
				}
				packet.Free()
			}
		}
	}()
}

func (in *Input) Close() {
	in.stop <- true
	in.wg.Wait()
	in.ctx.CloseInput()
}

func (in *Input) GetDataChannel() *utils.DataChannel {
	return &in.dataChannel
}

func (in *Input) GetMetaData() map[int]utils.MetaData {
	return in.metaData
}
