FROM golang:1.25.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build

EXPOSE 8080

# Run
CMD ["./discord"]