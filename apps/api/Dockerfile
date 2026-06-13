FROM golang:1.26

WORKDIR /app

COPY go.mod  ./
RUN go mod download

COPY ./ ./

RUN ls cmd


RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd ./cmd/main.go 

RUN ls

CMD ["/cmd"]

