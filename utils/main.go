package utils

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"pi-thermo-guard/constants"
	"strings"
	"syscall"
	"time"
)

func NewServer(addr string) *http.Server {
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		ErrorLog:     log.New(os.Stderr, "HTTP: ", 0),
	}
}

func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		select {
		case <-sig:
			cancel()
		}
	}()

	return ctx
}

func SendEmail(to []string, disableShutdown string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("EMAIL_HOST")
	hostPort := os.Getenv("EMAIL_HOST_PORT")
	if from == "" || password == "" || host == "" || hostPort == "" {
		return fmt.Errorf("Environment variables for sender email are missing")
	}

	fullHostAddress := host + ":" + hostPort

	mailInfoHeader := "From: " + constants.EmailFromName + " <" + from + ">\r\nTo: " + strings.Join(to, ", ") + "\r\nSubject: " + constants.EmailSubject + "\r\n"
	mimeHeader := "MIME: MIME-version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"

	message := fmt.Sprintf("Timestamp: %s\nThe Raspberry Pi (%s <%s>) was overheating!\n\nReaction: Program exited", time.Now().Format(time.DateTime), GetHostname(), GetLocalIP())
	if disableShutdown != "true" {
		message += " + Shutdown performed"
	}

	mail := []byte(mailInfoHeader + mimeHeader + message)
	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(fullHostAddress, auth, from, to, mail)
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	} else {
		return ""
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
