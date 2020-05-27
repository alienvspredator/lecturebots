FROM golang:latest AS builder
ARG sourcedir=$GOPATH/src/github.com/alienvspredator/tgbot
ARG builddir=/bot

WORKDIR ${sourcedir}

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make bot
RUN mkdir ${builddir}
RUN mv build/bot ${builddir}/bot
RUN rm -rf ${sourcedir}

FROM builder as runner

ENTRYPOINT [ "/bot/bot" ]
