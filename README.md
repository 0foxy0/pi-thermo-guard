# PiThermoGuard
A temperature monitoring program with Prometheus metrics and email notifications for the Raspberry Pi!\
Created using Golang.
___

## About
This program monitors the temperature of the Raspberry Pi's CPU (every 15 seconds) to avoid overheating and help creating Metrics about it.\
So when the temperature limit which you can configure gets exceeded the RASPI gets shut down if not disabled and sends an email to the specified contacts!\
Also the monitored temperatures are being available for your prometheus in the web api (0.0.0.0:PORT/metrics).\
You can deploy it using docker or use your own approach building it yourself as you can see at [Build](#build)!

## Build
### Docker
> see [Dockerfile](./Dockerfile)
```
volumes:
  ptg-data:
    driver: local

services:
  pi_thermo_guard:
    image: 0foxy0/pi-thermo-guard:latest
    env_file:
      - .env
    container_name: pi_thermo_guard
    ports:
      - "4440:4440" # has to match the "PORT" of .env
    volumes:
      - ptg-data:/app
```

### build command for local installation
`env GOOS=linux GOARCH=arm go build -o PiThermoGuard`

## Environment Variables
> see [.env.example](./.env.example)
- PORT (port for the prometheus web api)
- TEMPERATURE_LIMIT (unit: Celsius, type: float64)
- NOTIFICATION_EMAILS (The specified emails get send the notification when the raspberry pi overheats) [Optional]
- EMAIL_ADDRESS [Optional]
- EMAIL_PASSWORD [Optional]
- EMAIL_HOST [Optional]
- EMAIL_HOST_PORT [Optional]
- DISABLE_SHUTDOWN [Optional]
