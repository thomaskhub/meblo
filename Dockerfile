FROM golang:1.21 as base

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -y \
    autoconf \
    automake \
    build-essential \
    cmake \
    git-core \
    libass-dev \
    libfreetype6-dev \
    libunistring-dev \
    libgnutls28-dev \
    libmp3lame-dev \
    libsdl2-dev \
    libtool \
    libva-dev \
    libvdpau-dev \
    libvorbis-dev \
    libxcb1-dev \
    libxcb-shm0-dev \
    libxcb-xfixes0-dev \
    libgnutls28-dev \
    meson \
    ninja-build \
    pkg-config \
    texinfo \
    lzma \
    wget \
    yasm \
    zlib1g-dev \
    ca-certificates \
    libssl-dev

#Setup and Install paho library
WORKDIR /source
RUN git clone https://github.com/eclipse/paho.mqtt.c.git paho \
    && cd paho \
    && git checkout v1.3.12 \
    && make \
    && make install \
    && rm -rf /usr/local/share/man

RUN wget https://www.nasm.us/pub/nasm/releasebuilds/2.15.05/nasm-2.15.05.tar.bz2 \
    && tar xjvf nasm-2.15.05.tar.bz2 \
    && cd nasm-2.15.05 \
    && ./autogen.sh \
    && ./configure \
    && make \
    && make install \
    && rm -rf /usr/local/share/man

RUN git -C x264 pull 2>/dev/null || git clone --depth 1 https://code.videolan.org/videolan/x264.git \
    && cd x264 \
    && ./configure --enable-static --enable-pic \
    && make \
    && make install

RUN git clone https://github.com/thomaskhub/FFmpeg.git ffmpeg-6.0 \
    && cd ffmpeg-6.0 \
    && git checkout ums \
    && ./configure --extra-libs="-lpthread -lm" --ld="g++" --enable-gpl --enable-gnutls --enable-libx264 --extra-libs="-lpthread" \
    && make \
    && make install \
    && rm -rf /usr/local/share/man

RUN apt-get install libjansson-dev -y

FROM base as dev

RUN apt-get install -y gdb

FROM base as prod
RUN rm -rf /source/*
COPY ./ /source/ums/
WORKDIR /source/ums
RUN make && chmod ugo+x ./ums
CMD ["/source/ums/ums"]

# ENV PKG_CONFIG_PATH=/ffmpeg_build/lib/pkgconfig

# FROM base as devel
# RUN /usr/bin/make debug

# FROM base as prod
# RUN /usr/bin/make

# CMD ["./ums"]
