package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"os/exec"
	"pi-thermo-guard/constants"
	"pi-thermo-guard/utils"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := utils.Context()
	errorLog := log.New(os.Stderr, "PTG: ", 0)

	currentWd, _ := os.Getwd()
	if err := godotenv.Load(currentWd + "/.env"); err != nil {
		fmt.Println("No .env file found")
	}

	tempLimitAsStr := os.Getenv("TEMPERATURE_LIMIT")
	if tempLimitAsStr == "" {
		errorLog.Println("Environment variable TEMPERATURE_LIMIT is missing")
		return 1
	}

	tempLimit, err := strconv.ParseFloat(tempLimitAsStr, 64)
	if err != nil {
		errorLog.Printf("Couldn't parse TEMPERATURE_LIMIT! %s", tempLimitAsStr)
		return 1
	}

	port := os.Getenv("PORT")
	if port == "" {
		errorLog.Println("Environment variable PORT is missing")
		return 1
	}

	toEmails := os.Getenv("NOTIFICATION_EMAILS")
	disableShutdown := os.Getenv("DISABLE_SHUTDOWN")

	server := utils.NewServer(":" + port)

	tempGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "pi_cpu",
		Name:      "temperature",
		Help:      "The CPU temperature of this Raspberry Pi in Celsius degrees.",
	})
	prometheus.MustRegister(tempGauge)

	go func() {
		code := measureTemp(ctx, tempLimit, tempGauge, toEmails, disableShutdown, errorLog)
		if code != 0 {
			os.Exit(code)
		} else {
			syscall.Kill(syscall.Getpid(), syscall.SIGQUIT)
		}
	}()
	go func() { log.Fatal(server.ListenAndServe()) }()

	<-ctx.Done()
	err = server.Shutdown(utils.Context())
	if err != nil {
		return 1
	}

	return 0
}

func measureTemp(ctx context.Context, tempLimit float64, tempGauge prometheus.Gauge, toEmails string, disableShutdown string, errorLog *log.Logger) int {
	for {
		buffer, err := os.ReadFile(constants.TempPath)
		if err != nil {
			errorLog.Printf("Couldn't read temp file! %s", constants.TempPath)
			return 1
		}

		fileContent := strings.TrimSpace(string(buffer))

		millC, err := strconv.ParseFloat(fileContent, 64)
		if err != nil {
			errorLog.Printf("Couldn't parse temp file content! '%s'", fileContent)
			return 1
		}

		temp := millC / 1000
		tempGauge.Set(temp)

		if temp > tempLimit {
			if disableShutdown != "true" {
				cmd := exec.Command("sudo", constants.ShutdownCmd...)
				err := cmd.Run()
				if err != nil {
					errorLog.Printf("Couldn't shutdown the Raspberry Pi! (used command: sudo %s)", strings.Join(constants.ShutdownCmd, " "))
					return 1
				}
			}

			if toEmails != "" {
				if utils.SendEmail(strings.Split(toEmails, ", "), disableShutdown) != nil {
					errorLog.Printf("Couldn't send a email to the following: %s", toEmails)
				}
			}

			errorLog.Println("Raspberry Pi overheats!")
			return 0
		}

		select {
		case <-time.After(time.Second * 15):
			continue
		case <-ctx.Done():
			return 0
		}
	}
}
