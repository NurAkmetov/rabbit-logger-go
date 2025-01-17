package rabbit_logger_go

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Protocol    string
	Timezone    string
	Hostname    string
	Port        int
	Username    string
	Password    string
	VHost       string
	Queue       string
	Env         string
	ProjectName string
}

type LogMessage struct {
	ActionName    string `json:"actionName,omitempty"`
	ActionStage   string `json:"actionStage,omitempty"`
	RequestID     string `json:"requestId,omitempty"`
	Message       string `json:"message"`
	TransactionID string `json:"transactionId,omitempty"`
	Context       string `json:"context,omitempty"`
	Response      string `json:"response,omitempty"`
	Backtrace     string `json:"backtrace,omitempty"`
}

type DefaultConf struct {
	Date        string `json:"date"`
	DateTime    string `json:"dateTime"`
	Timestamp   int64  `json:"timestamp"`
	Environment string `json:"environment"`
	ProjectName string `json:"projectName"`
	LogLevel    string `json:"logLevel"`
}

type RabbitLogger struct {
	Config     Config
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func NewRabbitLogger(config Config) (*RabbitLogger, error) {
	if config.Protocol == "" {
		config.Protocol = "amqp"
	}
	if config.Timezone == "" {
		config.Timezone = "Asia/Aqtau"
	}

	conn, err := amqp.Dial(fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		config.Protocol, config.Username, config.Password, config.Hostname, config.Port, config.VHost))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	if _, err := ch.QueueDeclare(config.Queue, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	logger := &RabbitLogger{
		Config:     config,
		Connection: conn,
		Channel:    ch,
	}
	log.Println("Connected to RabbitMQ (ClickHouse Logger).")
	return logger, nil
}

func (r *RabbitLogger) Close() {
	if err := r.Channel.Close(); err != nil {
		log.Println("Failed to close RabbitMQ channel:", err)
	}
	if err := r.Connection.Close(); err != nil {
		log.Println("Failed to close RabbitMQ connection:", err)
	}
}

func (r *RabbitLogger) getDefaultConfig(logLevel string) DefaultConf {
	loc, _ := time.LoadLocation(r.Config.Timezone)
	now := time.Now().In(loc)

	return DefaultConf{
		Date:        now.Format("2006-01-02"),
		DateTime:    now.Format("2006-01-02 15:04:05"),
		Timestamp:   now.Unix(),
		Environment: r.Config.Env,
		ProjectName: r.Config.ProjectName,
		LogLevel:    logLevel,
	}
}

func (r *RabbitLogger) sendToQueue(logLevel string, data LogMessage) error {
	defaultConf := r.getDefaultConfig(logLevel)
	payload := map[string]interface{}{
		"date":        defaultConf.Date,
		"dateTime":    defaultConf.DateTime,
		"timestamp":   defaultConf.Timestamp,
		"environment": defaultConf.Environment,
		"projectName": defaultConf.ProjectName,
		"logLevel":    defaultConf.LogLevel,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return fmt.Errorf("failed to merge data into payload: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal final payload: %w", err)
	}

	return r.Channel.Publish(
		"",             // exchange
		r.Config.Queue, // routing key (queue name)
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payloadBytes,
		},
	)
}

func (r *RabbitLogger) Info(data LogMessage) {
	if err := r.sendToQueue("INFO", data); err != nil {
		log.Printf("Failed to send log to queue: %v", err)
	}
}

func (r *RabbitLogger) Error(data LogMessage) {
	if err := r.sendToQueue("Error", data); err != nil {
		log.Printf("Failed to send log to queue: %v", err)
	}
}

func (r *RabbitLogger) Warning(data LogMessage) {
	if err := r.sendToQueue("Warning", data); err != nil {
		log.Printf("Failed to send log to queue: %v", err)
	}
}

func (r *RabbitLogger) Debug(data LogMessage) {
	if err := r.sendToQueue("DEBUG", data); err != nil {
		log.Printf("Failed to send log to queue: %v", err)
	}
}
