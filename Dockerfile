FROM golang:1.22

WORKDIR /build

COPY . .
RUN go build -o flywatch .

CMD ["./flywatch"]

