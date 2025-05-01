FROM golang:1.24

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOD=linux go build -o /konkurransetilsynet

CMD ["/konkurransetilsynet"]
