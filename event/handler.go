package event

import (
	"EventTrigger/data"
	"log"
	"time"
)

type Request struct {
	ExecutionType string
	TriggerType   string
	UserId        uint64
	TriggerId     uint64
	API           string
	Payload       string
}

func (r *Request) HandleRequest() {
	go func() {

		err := data.DB.Create(data.Event{
			Type:      r.TriggerType,
			Timestamp: time.Now(),
			Execution: r.ExecutionType,
			UserId:    r.UserId,
			TriggerId: r.TriggerId,
			Payload:   r.Payload,
		}).Error

		if err != nil {
			log.Println("Failed log event reason ", err.Error())
		}
	}()
}
