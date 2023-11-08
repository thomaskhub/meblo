# meblo

Open Media Blocks - A collection of Golang code to create media streaming services the way you need it

The intention of this library is to create a catalog of moudles that can be used to setup customized media
streaming servers example are:

- RTMP stream multiplication to stream one RTMP stream to many servers
- DASH / HLS media streaming
- Creation of VODs with DRM technologies

## meblo elements

![Medie pip](./doc/img/pipe.png)

In meblo we can use different types of components to create our media streaming pipeline.

**Inputs** modules will read data from input which can be files like video files or png files, or it can be network
based protocols like RTMP or UDP. Each Input elements will
provide to channels of type \*AVPacket which can be used
by the subsequent pipe elements. The next component in the
pipe must ensure to clean packet memory.

**Decoder** will take the packets from the inputs and
decode them into their raw format. The output of the decoder
will be audio and video channels of type \*AVFrame. The
next element in the pipe needs to ensure to clean frame
memory once frame is not needed any longer

**Filter** will be able to take the AVFrame object and
process them . The output of these modules will be channels
of type \*AVFrame. The next stage is responsible for freeing
memory for AVFrames on the outputs.

**Encoder** from filters it will go into encoder that
will take the raw image and encode them with respective
configurations like resolution and bitrate. The input
of encoder will be channels of *AVFrame where as the output
will be channels of type *AVPacket.

**Outputs** will take the AVPackets from the Encoders and
convert it into the respectice output path. Outputs will
not have any channels as output as the data is directly saved
on disk or send over some network protocol

### Inputs

Inputs elemets can be used to read video data from different types of sources.
All inputs elements must follow the following interface

```golang
type {InputName}Input struct {
	base           BaseInput        //the base input
	videoOut       chan *C.AVPacket //video data
	audioOut       chan *C.AVPacket //audio data
	audioTimescale utils.Timescale  //define the audio timescale
	videoTimescale utils.Timescale  //define the video timescale
}

func (in *{InputName}) Open() {} //open the input and start processing
func (in *{InputName}) Close() {} //stop the input and free all resources
func (in *{InputName}) GetStata() {} //return the internl statistics of the module
```

#### FileInput

Used to read video data from a file and output the data as \*AVPAckets.
The output buffers of this module are set to 16 elements in size. The path
of the file passed to the open function must be an absolute path.

```golang
fileInput := FileInput{}
err := fileInput.Open(absPath)

// use fileInput.videoOut and fileInput.audioOut to connect the input to the next pipe element

switch errorCode {
case utils.FFmpegErrorEinval:
    // handle internal error, some memory could not be allocated
case utils.FFmpegErrorEof:
    // handle the file has ended so we can close the input module
    fileInput.Close()
default:
    // handle other cases
}
```

#### RtmpPull Input (TODO)

- create an RTMP input module
- have an option in the open function that will allow to enable / disable retry
  - we should retry when connection is interrupted for what ever reason
  - reconnect to rtmp whenever
