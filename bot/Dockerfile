FROM golang:1.9.1-stretch

RUN apt-get update
RUN apt-get install -y ffmpeg
RUN apt-get install -y golang-glide

WORKDIR /go/src/app

COPY glide.yaml ./

RUN glide install

COPY . .

RUN go build -race -o capibot .
CMD ["./capibot"]
