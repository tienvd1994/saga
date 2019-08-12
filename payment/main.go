package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"log"
	"syreclabs.com/go/faker"
	"time"
)

const (
	PaymentChannel string = "PaymentChannel"
	ReplyChannel   string = "ReplyChannel"

	ActionStart    string = "Start"
	ActionDone     string = "DoneMsg"
	ActionError    string = "ErrorMsg"
	ActionRollback string = "RollbackMsg"
)

type Message struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

func (m Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func main() {
	var err error
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	if _, err = client.Ping().Result(); err != nil {
		log.Fatalf("error creating redis client %s", err)
	}

	pubsub := client.Subscribe(PaymentChannel, ReplyChannel)
	if _, err = pubsub.Receive(); err != nil {
		log.Fatalf("error subscribing %s", err)
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Println("starting the payment service")
	for {
		select {
		case msg := <-ch:
			m := Message{}
			err := json.Unmarshal([]byte(msg.Payload), &m)
			if err != nil {
				log.Println(err)
				continue
			}

			switch msg.Channel {
			case PaymentChannel:
				log.Printf("recieved message with id %s ", m.ID)
				d := faker.RandomInt(1, 3)
				time.Sleep(time.Duration(d) * time.Second)

				if m.Action == ActionStart {
					m.Action = ActionDone
					if err = client.Publish(ReplyChannel, m).Err(); err != nil {
						log.Printf("error publishing done-message to %s channel", ReplyChannel)
					}
					log.Printf("done message published to channel :%s", ReplyChannel)
				}

				if m.Action == ActionRollback {
					log.Printf("rolling back transaction with ID :%s", m.ID)
				}

			}
		}
	}
}
