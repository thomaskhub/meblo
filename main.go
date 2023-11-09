package main

import (
	"github.com/thomaskhub/meblo/ffmpeg"
)

func main() {
	ctx := ffmpeg.AVFormatContext{}
	ctx.AVFormatAllocContext()

	// ctx.AVFormatOpenInput("/home/thomas/Videos/002.mp4")
	// // ctx.AVFormatFindStreamInfo()
	// ctx.AVDumpFormat()

}
