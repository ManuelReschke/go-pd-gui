FROM golang:1.19

RUN apt-get update && apt-get -y upgrade
RUN apt-get -y install gcc
RUN apt-get -y install libgl1-mesa-dev
RUN apt-get -y install xorg-dev

#COPY . /go/src/app

ADD . /usr/src/app

WORKDIR /usr/src/app

RUN go mod download && go mod verify

#RUN fyne-cross linux
#RUN go build -o main .

#CMD ["./main"]