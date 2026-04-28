FROM golang:alpine
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/main.go
CMD ["./main"]