package main

import (
	"github.com/NurAkmetov/rabbit-logger-go"
	"log"
)

func main() {
	config := rabbit_logger_go.Config{
		Protocol:    "amqp",
		Timezone:    "Asia/Aqtau",
		Hostname:    "localhost",
		Port:        5672,
		Username:    "guest",
		Password:    "guest",
		VHost:       "/",
		Queue:       "log-queue",
		Env:         "development",
		ProjectName: "ExampleProject",
	}

	rabbitLogger, err := rabbit_logger_go.NewRabbitLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitLogger: %v", err)
	}
	defer rabbitLogger.Close()

	rabbitLogger.Info(rabbit_logger_go.LogMessage{
		ActionName:    "TestAction",
		ActionStage:   "success",
		TransactionID: "12345",
		Message:       "Test message",
	})
}
