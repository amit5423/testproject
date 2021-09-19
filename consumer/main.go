package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	db "consumer/db"
	util "consumer/util"

	kafka "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	EmpId string
	Name  string
	Dept  string
	Time  string
}

func main() {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(context.Background())

	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		fmt.Println("Got signal: ", sig)
		cancel()
	}()

	bootstrapServers := strings.Split(util.GetEnv(util.BootstrapServers, "localhost:9092"), ",")
	topic := util.GetEnv(util.Topic, "event")
	groupID := util.GetEnv(util.GroupID, "my-group")

	config := kafka.ReaderConfig{
		Brokers: bootstrapServers,
		GroupID: groupID,
		Topic:   topic,
		MaxWait: 500 * time.Millisecond}

	r := kafka.NewReader(config)

	fmt.Println("Consumer configuration: ", config)

	defer func() {
		err := r.Close()
		if err != nil {
			fmt.Println("Error closing consumer: ", err)
			return
		}
		fmt.Println("Consumer closed")
	}()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		fmt.Printf("Received message from %s-%d [%d]: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		var event Event
		json.Unmarshal([]byte(m.Value), &event)

		var document interface{}

		document = bson.D{
			{"EmpId", event.EmpId},
			{"Name", event.Name},
			{"Dept", event.Dept},
			{"Time", event.Time},
		}

		db.Savedata(document)
	}
}
