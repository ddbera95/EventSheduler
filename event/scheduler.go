package event

import (
	"EventTrigger/data"
	"github.com/emirpasic/gods/trees/binaryheap"
	"log"
	"time"
)

var TriggerScheduler = NewScheduler()

func Init() {

	StartRetentionJob()
	TriggerScheduler.Start()

	var triggers []data.Trigger
	if err := data.DB.Find(&triggers).Error; err != nil {
		log.Fatalf("failed to fetch triggers , reason %v", err)
		return
	}

	for _, trigger := range triggers {
		TriggerScheduler.AddEvent(Trigger{
			TriggerId: trigger.ID,
			UserId:    trigger.UserID,
			Duration:  trigger.Duration.Duration,
			Ticker:    false,
		})
	}
}

type Trigger struct {
	NextTimestamp time.Time
	Duration      time.Duration
	UserId        uint64
	TriggerId     uint64
	Ticker        bool
}

type Scheduler struct {
	heap            *binaryheap.Heap
	Receiver        chan Trigger
	Delete          chan uint64
	deletedTriggers map[uint64]struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		heap: binaryheap.NewWith(func(a, b interface{}) int {

			if a.(Trigger).NextTimestamp == b.(Trigger).NextTimestamp {
				return 0
			}

			if a.(Trigger).NextTimestamp.Before(b.(Trigger).NextTimestamp) {
				return -1
			}

			return 1
		}),
		Receiver:        make(chan Trigger),
		Delete:          make(chan uint64),
		deletedTriggers: make(map[uint64]struct{}),
	}
}

func (s *Scheduler) AddEvent(event Trigger) {
	event.NextTimestamp = time.Now().Add(event.Duration)
	s.Receiver <- event
}

func (s *Scheduler) DeleteEvent(triggerId uint64) {
	s.Delete <- triggerId
}

func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(time.Second)

		for {

			select {

			case event := <-s.Receiver:
				s.heap.Push(event)

			case triggerId := <-s.Delete:
				s.deletedTriggers[triggerId] = struct{}{}

			case <-ticker.C:
				s.execute()

			}
		}
	}()

}

func (s *Scheduler) execute() {
	now := time.Now()

	for {
		event, exists := s.heap.Peek()
		if !exists {
			return
		}

		currentEvent := event.(Trigger)

		if _, ok := s.deletedTriggers[currentEvent.TriggerId]; ok {
			delete(s.deletedTriggers, currentEvent.TriggerId)
			continue
		}

		if currentEvent.NextTimestamp.After(now) {
			return
		}

		s.heap.Pop()

		processEvent(currentEvent)

		if currentEvent.Ticker {
			currentEvent.NextTimestamp = currentEvent.NextTimestamp.Add(currentEvent.Duration)
			s.heap.Push(currentEvent)
		}
	}

}

func processEvent(event Trigger) {

	var triggerType string

	if event.Ticker {
		triggerType = "ticker"
	} else {
		triggerType = "timer"
	}

	reqeust := Request{
		ExecutionType: "triggered",
		UserId:        event.UserId,
		TriggerId:     event.TriggerId,
		TriggerType:   triggerType,
	}

	reqeust.HandleRequest()
}
