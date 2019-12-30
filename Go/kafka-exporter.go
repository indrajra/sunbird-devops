package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
)

const (
	zookeeperConn = "0.0.0.0:2181"
	cgroup        = "Test1"
	topic1        = "sunbirddev.analytics_metrics"
	topic2        = "sunbirddev.pipeline_metrics"
)

var msg []string

func main() {
	// setup sarama log to stdout
	sarama.Logger = log.New(os.Stdout, "", log.Ltime)
	http.HandleFunc("/", serve)

	// init consumer
	cg, err := initConsumer()
	if err != nil {
		fmt.Println("Error consumer goup: ", err.Error())
		os.Exit(1)
	}
	defer cg.Close()
	// run consumer
	go consume(cg)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func serve(w http.ResponseWriter, r *http.Request) {
	for _, value := range msg {
		fmt.Fprintf(w, "Value: %q", string(value))
	}

}

func initConsumer() (*consumergroup.ConsumerGroup, error) {
	// consumer config
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	config.Offsets.ProcessingTimeout = 10 * time.Second

	// join to consumer group
	cg, err := consumergroup.JoinConsumerGroup(cgroup, []string{topic1, topic2}, []string{zookeeperConn}, config)
	if err != nil {
		return nil, err
	}
	return cg, err
}

func consume(cg *consumergroup.ConsumerGroup) {
	for {
		select {
		case message := <-cg.Messages():
			msg = append(msg, string(message.Value))
			err := cg.CommitUpto(message)
			if err != nil {
				fmt.Println("Error commit zookeeper: ", err.Error())
			}
		}
	}
}
