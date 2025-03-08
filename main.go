package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v3"
)

// Config structure for YAML
type Config struct {
	RabbitMQ struct {
		Host   string   `yaml:"host"`
		Port   int      `yaml:"port"`
		User   string   `yaml:"user"`
		Pwd    string   `yaml:"pwd"`
		Queues []string `yaml:"queues"`
	} `yaml:"rabbitmq"`
	Metrics struct {
		Host     string `yaml:"host"`     // Bind address (127.0.0.1 or 0.0.0.0)
		Port     int    `yaml:"port"`     // Prometheus metrics port
		Interval int    `yaml:"interval"` // Metrics update interval
	} `yaml:"metrics"`
}

// LoadConfig reads config from YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Prometheus metrics
var (
	queueConsumers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_queue_consumers",
			Help: "Number of active consumers per queue",
		},
		[]string{"queue"},
	)

	queueMessages = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_queue_messages",
			Help: "Total number of messages in the queue",
		},
		[]string{"queue"},
	)
)

func main() {
	// Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Register Prometheus metrics
	prometheus.MustRegister(queueConsumers)
	prometheus.MustRegister(queueMessages)

	// Get bind address from config
	metricsAddr := fmt.Sprintf("%s:%d", config.Metrics.Host, config.Metrics.Port)

	// Start HTTP server for Prometheus
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Prometheus metrics available at http://%s/metrics", metricsAddr)
		log.Fatal(http.ListenAndServe(metricsAddr, nil))
	}()

	// Construct RabbitMQ connection URL
	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.RabbitMQ.User, config.RabbitMQ.Pwd,
		config.RabbitMQ.Host, config.RabbitMQ.Port,
	)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Get update interval from config
	updateInterval := time.Duration(config.Metrics.Interval) * time.Second
	log.Printf("Queue monitoring interval set to %d seconds", config.Metrics.Interval)

	// Monitor queues in a loop
	for {
		log.Println("Updating queue metrics...")

		for _, queueName := range config.RabbitMQ.Queues {
			queue, err := ch.QueueInspect(queueName)
			if err != nil {
				log.Printf("Failed to inspect queue '%s': %v", queueName, err)
			} else {
				queueConsumers.WithLabelValues(queueName).Set(float64(queue.Consumers))
				queueMessages.WithLabelValues(queueName).Set(float64(queue.Messages))

				log.Printf("Queue '%s': consumers=%d, messages=%d",
					queueName, queue.Consumers, queue.Messages)
			}
		}

		time.Sleep(updateInterval)
	}
}
