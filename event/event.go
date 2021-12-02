package event

import "sync"

type EventChannel chan interface{}

var (
	// A collection of subscribers to different events.
	subscribers = map[string][]EventChannel{}
	mutex       sync.Mutex
)

// Subscribe will add the event channel to the corresponding table.
func Subscribe(topic string, ch EventChannel) {
	mutex.Lock()
	if list, found := subscribers[topic]; found {
		subscribers[topic] = append(list, ch)
	} else {
		// if this topic has no subscriber
		subscribers[topic] = append([]EventChannel{}, ch)
	}
	mutex.Unlock()
}

// Publish is responsible for pushing the data of the event to the specified
// kind of subscribers, call it will put an empty struct in it as usual.
func Publish(topic string, event interface{}) {
	mutex.Lock()
	if list, found := subscribers[topic]; found {
		go func(event interface{}, list []EventChannel) {
			for _, subscriber := range list {
				subscriber <- event
			}
		}(event, append([]EventChannel{}, list...))
	}
	mutex.Unlock()
}
