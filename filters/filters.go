package filters

import (
	"errors"
	"strconv"

	"github.com/thomaskhub/go-astiav"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/utils"
)

type Filter struct {
	srcFilter   *astiav.Filter
	sinkFilter  *astiav.Filter
	srcCtx      *astiav.FilterContext
	sinkCtx     *astiav.FilterContext
	filterGraph *astiav.FilterGraph
	filterIn    *astiav.FilterInOut
	filterOut   *astiav.FilterInOut
	filterFrame *astiav.Frame

	dataChannel    chan astiav.Frame
	metaData       utils.MetaData
	dataChannelOut chan astiav.Frame
	metaDataOut    utils.MetaData
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) SetMetaData(metaData utils.MetaData) {
	f.metaData = metaData
}

func (f *Filter) SetDataChannel(data chan astiav.Frame) {
	f.dataChannel = data
}

func (f *Filter) GetDataChannel() chan astiav.Frame {
	return f.dataChannelOut
}

func (f *Filter) GetMetaData() utils.MetaData {
	return f.metaDataOut
}

// Added frame rate to filter as I was not able to figure out how to pass new frame rate to the output
// meta data in case the frame rate is adjusted by the filter
func (f *Filter) OpenVideo(filter string, outWidth int, outHeight int, frameRate *astiav.Rational) {
	width := f.metaData.CodecPar.Width()
	height := f.metaData.CodecPar.Height()

	args := astiav.FilterArgs{
		"pixel_aspect": f.metaData.CodecPar.SampleAspectRatio().String(),
		"pix_fmt":      f.metaData.CodecPar.PixelFormat().String(),
		"time_base":    f.metaData.TimeBase.String(),
		"video_size":   strconv.Itoa(width) + "x" + strconv.Itoa(height),
	}

	f.metaDataOut = f.metaData
	f.metaDataOut.CodecPar.SetWidth(outWidth)
	f.metaDataOut.CodecPar.SetHeight(outHeight)

	f.open(filter, "buffer", "buffersink", args, frameRate)
}

func (f *Filter) open(filter string, srcName string, sinkName string, args astiav.FilterArgs, frameRate *astiav.Rational) {
	var err error

	f.dataChannelOut = make(chan astiav.Frame, 16)

	f.filterGraph = astiav.AllocFilterGraph()
	if f.filterGraph == nil {
		logger.Fatal("could not alloc filter graph")
	}

	f.filterIn = astiav.AllocFilterInOut()
	if f.filterIn == nil {
		logger.Fatal("could not alloc filter in")
	}

	f.filterOut = astiav.AllocFilterInOut()
	if f.filterOut == nil {
		logger.Fatal("could not alloc filter out")
	}

	f.srcFilter = astiav.FindFilterByName(srcName)
	if f.srcFilter == nil {
		logger.Fatal("could not find source filter")
	}

	f.sinkFilter = astiav.FindFilterByName(sinkName)
	if f.sinkFilter == nil {
		logger.Fatal("could not find sink filter")
	}

	f.srcCtx, err = f.filterGraph.NewFilterContext(f.srcFilter, "in", args)
	if err != nil {
		logger.Fatal("could not create source filter context")
	}

	f.sinkCtx, err = f.filterGraph.NewFilterContext(f.sinkFilter, "out", nil)
	if err != nil {
		logger.Fatal("could not create sink filter context")
	}

	f.filterOut.SetName("in")
	f.filterOut.SetFilterContext(f.srcCtx)
	f.filterOut.SetPadIdx(0)
	f.filterOut.SetNext(nil)

	f.filterIn.SetName("out")
	f.filterIn.SetFilterContext(f.sinkCtx)
	f.filterIn.SetPadIdx(0)
	f.filterIn.SetNext(nil)

	err = f.filterGraph.Parse(filter, f.filterIn, f.filterOut)
	if err != nil {
		logger.Fatal("could not parse filter graph")
	}

	err = f.filterGraph.Configure()
	if err != nil {
		logger.Fatal("could not configure filter graph")
	}

	f.filterFrame = astiav.AllocFrame()
	f.metaDataOut.TimeBase = f.sinkCtx.Inputs()[0].TimeBase()

	if frameRate != nil {
		f.metaDataOut.FrameRate = *frameRate

	}

}

func (f *Filter) Run() {

	go func() {
		for {

			frame := <-f.dataChannel

			err := f.srcCtx.BuffersrcAddFrame(frame.Clone(), astiav.NewBuffersrcFlags(astiav.BuffersrcFlagPush))
			if err != nil {
				logger.Fatal("could not add frame")
			}

			frame.Unref()

			err = f.sinkCtx.BuffersinkGetFrame(f.filterFrame, astiav.NewBuffersinkFlags())
			if err != nil {
				if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
					err = nil
					continue
				}
				logger.Fatal("could not get frame")
			}

			f.dataChannelOut <- *f.filterFrame.Clone()
			f.filterFrame.Unref()
		}
	}()
}
