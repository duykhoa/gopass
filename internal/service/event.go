package service

type EventType string

const (
	EventError   EventType = "error"
	EventSuccess EventType = "success"
	EventInfo    EventType = "info"
)

type Event struct {
	Type    EventType
	Message string
	Data    interface{}
}

type Subscriber func(Event)

type PubSub struct {
	subscribers []Subscriber
}

func NewPubSub() *PubSub {
	return &PubSub{}
}

func (ps *PubSub) Subscribe(sub Subscriber) {
	ps.subscribers = append(ps.subscribers, sub)
}

func (ps *PubSub) Publish(event Event) {
	for _, sub := range ps.subscribers {
		sub(event)
	}
}
