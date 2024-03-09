FROM golang:1.20

WORKDIR /app

COPY . ./

RUN go mod tidy
RUN env GOOS=linux GOARCH=arm go build -o PiThermoGuard

ENTRYPOINT ["/app/PiThermoGuard"]