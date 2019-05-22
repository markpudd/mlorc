FROM golang:1.10

WORKDIR $GOPATH/src/github.com/markpudd/mlorc

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

EXPOSE 8585

CMD ["mlorc"]
