package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	irc "github.com/fluffle/goirc/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type conf struct {
	Server           string   `yaml:"server"`
	Port             int64    `yaml:"port"`
	UseSSL           bool     `yaml:"useSSL"`
	Channels         []string `yaml:"channels"`
	Nick             string   `yaml:"nick"`
	BootstrapServers string   `yaml:"bootstrapServers"` // Connection string for Kafka cluster
	TopicPrefix      string   `yaml:"topicPrefix"`      // Topic prefix; topics will be made in the form {prefix}_{channel}
}

type message struct {
	Channel string
	Nick    string
	Text    string
	Time    int64
}

var p *kafka.Producer

func (c *conf) getConf() *conf {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func bootstrapKafkaProducer(cfg *kafka.ConfigMap) {
	if p == nil {
		var err error
		p, err = kafka.NewProducer(cfg)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	var beaverConf conf
	beaverConf.getConf()
	log.Printf("Loaded configuration for Beaver!")
	blob, _ := yaml.Marshal(beaverConf)
	log.Printf("Using configuration: \n%s", string(blob))

	log.Printf("Starting Kafka producer")
	bootstrapKafkaProducer(&kafka.ConfigMap{"bootstrap.servers": beaverConf.BootstrapServers})
	defer p.Close()
	log.Printf("Created Kafka producer")

	cfg := irc.NewConfig(beaverConf.Nick)
	cfg.SSL = beaverConf.UseSSL
	cfg.SSLConfig = &tls.Config{ServerName: beaverConf.Server}
	cfg.Server = fmt.Sprintf("%s:%d", beaverConf.Server, beaverConf.Port)
	cfg.NewNick = func(n string) string { return n + "^" }
	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range beaverConf.Channels {
			log.Printf("Joining channel %s", channel)
			c.Join(channel)
		}
	})

	deliveryChan := make(chan kafka.Event)

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		if line.Public() {
			message := message{
				Channel: line.Target(),
				Nick:    line.Nick,
				Text:    line.Text(),
				Time:    line.Time.Unix(),
			}
			blob, _ := json.Marshal(message)
			log.Printf("Received message: %s", blob)
			topic := fmt.Sprintf("%s_%s", beaverConf.TopicPrefix, line.Target()[1:])
			log.Printf("Attempting to produce message to topic %s", topic)
			p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          []byte(blob),
			}, deliveryChan)

			e := <-deliveryChan
			m := e.(*kafka.Message)

			if m.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
			} else {
				log.Printf("Delivered message to topic %s [%d] at offset %v\n",
					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			}

		}
	})

	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})
	if err := c.Connect(); err != nil {
		log.Printf("Connection error: %s\n", err.Error())
	}
	<-quit
}
