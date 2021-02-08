FROM golang:1.15

RUN go version
ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN go build -o main .

CMD ["./main"]