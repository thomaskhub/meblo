# Why meblo

We wanted to have a simple way of creating different streaming applications
using ffmpeg library. FFMPEG is nice but C not necessarily the language that most
people are proficient in.

Gstreamer is also nice, but I felt its to complicated to get started.
Thats why we crated meblo which uses ffmpeg and provides some simple
media-blocks (meblo) to create different streaming applications.

It is based on (go-astiav)[https://github.com/asticode/go-astiav] library
but because it was not supporting FFMPEG 6 we forked it and add small changes
to support ffmpeg 6. We also added support for the FFMPEG Audio fifo.

The major task of meblo will be reading video data from various sources
like files, or rtmp servers and converting them to dash / hls /mpegts / rtmp formats.
We wante to have a switching system which allows switching between multiple
input streams without worrying about timestamp generation and possible gaps in the
content. The idea of meblo is this:

Take a video input (file or stream) and normalize it.
Normalizing means we fix the framrate for the video to 25fps and audio sample
rate to 44100Hz.

With 25fps one frame has a duration of 40ms. It turns out that 40ms of audio
is represented by exactly 1764 sample of audio @ 44100Hz. So we are clubbing
1 video frame with 1 audio frame, ensuring that they are in sync with in this 40ms.
We call the combination of 1 video frame and 1764 audio samples a **chunk**

Then switching between input streams becomes very easy without the need to worry
about gaps or thinking about the PTS timestamps. PTS time stamps just have to
be increased by 40ms with each chunk processed.

No gaps, no pts hassle, no worries !

Note: this is still under devlopment, not all the block are ready yet :)
Right now we have input, output, encryption, decryption and filter elements ready.
Next would be the setup of higher leverl meblos that create the **chunks** and
perform the switching between the streams. We are not so much focused on manipulating
the video data except changing resolutions and bitrates. But using the filer blocks
everything that can be done with ffmpeg can be done here also.

# Examples

## Remuxing

Takes in input video from the assets directory and writes it to disk as mpeg ts
under `/tmp/oranges.ts`

```bash
cd remuxing
go run .
```

## Codec

This example take the test video transcodes it and saves it under `/tmp/tomatoes.ts`

```bash
cd codec
go run .
```

## Chunker

This example reads a mp4 file and runs the audio stream through the
chunker to create audio frames with 1764 samples and then converts it back
to audio frames with 1024 samples. The output file is saved under /tmp/chunker.ts

Requirement:

- Need ums-rtmp docker container up and running
- Need to use ffmpeg to stream a v

```bash
cd chunker

ffmpeg -re -i ../../assets/test/test.mp4 -c copy -f flv rtmp://localhost:1935/live/test &
go run .
```
