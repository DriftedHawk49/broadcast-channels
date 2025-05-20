package broadcastchannels

import (
	"sync"

	"github.com/google/uuid"
)

type BroadcastInterface[T any] interface {
	/*Id is returned for getting listeners*/
	Subscribe() string
	/*
		This closes broadcast channels, effectively closing all listeners as well.
		Make sure you are consuming listeners after checking their aliveness.
	*/
	Close()
	/*
		Detaches a listener identified by id. non existent id is a no-op.
	*/
	Unsubscribe(id string)
	/*
		Broadcasts data to all the active listeners
	*/
	Broadcast(T)
	/*
		Get Listener by id. id is generated at the time of subscription.
		in case id does not exist, this will return a nil channel.
		Note : checking for nil-ness is important, as it will block forever if read from, unchecked.
	*/
	Listener(id string) <-chan T
}

type broadcastChannel[T any] struct {
	chans    sync.Map // map that will contain all the channels who have subscribed to this channel
	buffer   int      // buffer for channels
	isClosed bool     // tracks the request of closing of broadcast channel
}

// Use Id returned by this function to get particular listener
func (b *broadcastChannel[T]) Subscribe() string {
	id := uuid.New().String()
	b.chans.Store(id, make(chan T, b.buffer))
	return id
}

// Closes the channel
func (b *broadcastChannel[T]) Close() {

	if b.isClosed {
		return
	}

	b.chans.Range(func(key, value any) bool {
		c := value.(chan T)
		close(c)
		return true
	})

	b.isClosed = true
}

func (b *broadcastChannel[T]) Unsubscribe(id string) {
	v, ok := b.chans.LoadAndDelete(id)
	if !ok {
		return
	}

	close(v.(chan T))
}

func (b *broadcastChannel[T]) Broadcast(data T) {

	b.chans.Range(func(key, value any) bool {
		go func() {
			c, ok := value.(chan T)
			if ok {
				c <- data
			}
		}()
		return true
	})
}

// If there is no subscription with provided id, then it will return nil, hence nil check is necessary
func (b *broadcastChannel[T]) Listener(id string) <-chan T {

	v, ok := b.chans.Load(id)
	if !ok {
		return nil
	}
	return v.(chan T)
}

// Accepts optional buffer size for channel, if multiple buffers are provided
// , final buffer will be sum of all provided buffers
func NewBroadcastChannel[T any](buffer ...int) BroadcastInterface[T] {
	b := 0
	for _, buf := range buffer {
		b += buf
	}

	return &broadcastChannel[T]{
		buffer: b,
		chans:  sync.Map{},
	}
}
