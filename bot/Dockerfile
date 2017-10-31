FROM golang:1.9.1-stretch

RUN apt-get update
RUN apt-get install -y ffmpeg

RUN go get -u github.com/golang/dep/...

WORKDIR /go/src/app

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

COPY . .

RUN go build -race -o capibot .
CMD ["./capibot"]
