package broadcastchannels

import (
	"broadcast-channels/constants"
	"errors"
	"sync"

	"github.com/google/uuid"
)

/*
More functions for support in future :
1. Reset : Removes all listeners, probably for fresh re-use
*/
type BroadcastInterface[T any] interface {
	/*Id is returned for getting listeners, */
	Subscribe() (string, *Listener[T], error)
	/*
		This closes broadcast channels, effectively closing all listeners as well.
		Make sure you are consuming listeners after checking their aliveness.
		After closure of broadcast channel, object cannot be used again.
	*/
	Close()
	/*
		Detaches a listener identified by id. non existent id or closed channel is a no-op.
	*/
	Unsubscribe(id string)
	/*
		Broadcasts data to all the active listeners.
		Broadcast is blocking in nature, so make sure all the subscribed listners are
		listening on the channel and consuming it.

	*/
	Broadcast(T) error
}

type Listener[T any] struct {
	DoneChan <-chan struct{}
	DataChan <-chan T
}

type internalListener[T any] struct {
	doneChan chan struct{}
	dataChan chan T
}

type broadcastChannel[T any] struct {
	chans     sync.Map   // map that will contain all the channels who have subscribed to this channel
	buffer    int        // buffer for channels
	isClosed  bool       // tracks the request of closing of broadcast channel
	closeLock sync.Mutex // lock for operations
	singleOp  *sync.Once // for closing broadcast channel only once even if multiple process call it
}

/*
Returns Id, listener and error if any.
Id : A unique identifier for this particular subscription.
Listner which
*/
func (b *broadcastChannel[T]) Subscribe() (string, *Listener[T], error) {

	if b.isClosed {
		return "", nil, errors.New(constants.MessageClosedChannel)
	}

	b.closeLock.Lock()
	defer b.closeLock.Unlock()

	id := uuid.New().String()

	dataChan := make(chan T, b.buffer)
	doneChan := make(chan struct{})
	b.chans.Store(id, &internalListener[T]{
		doneChan: doneChan,
		dataChan: dataChan,
	})
	return id, &Listener[T]{
		DoneChan: doneChan,
		DataChan: dataChan,
	}, nil
}

// Closes the channel
func (b *broadcastChannel[T]) Close() {

	b.closeLock.Lock()
	defer b.closeLock.Unlock()
	if b.isClosed {
		return
	}
	b.isClosed = true

	// should be run only once even if race occurs to close the channel
	b.singleOp.Do(func() {
		b.chans.Range(func(key, value any) bool {
			c := value.(*internalListener[T])
			if c != nil {
				// inform for graceful shutdown
				go func() { c.doneChan <- struct{}{}; close(c.doneChan) }()
				// close the data channel
				close(c.dataChan)
			}
			return true
		})
	})

}

func (b *broadcastChannel[T]) Unsubscribe(id string) {
	b.closeLock.Lock()
	defer b.closeLock.Unlock()

	if b.isClosed {
		return
	}

	v, ok := b.chans.LoadAndDelete(id)
	if !ok {
		return
	}

	l := v.(*internalListener[T])
	if l == nil {
		return
	}

	// for graceful shutdown of listner
	l.doneChan <- struct{}{}

	close(l.doneChan)
	close(l.dataChan)
}

func (b *broadcastChannel[T]) Broadcast(data T) error {

	b.closeLock.Lock()
	defer b.closeLock.Unlock()

	if b.isClosed {
		return errors.New(constants.MessageClosedChannel)
	}

	b.chans.Range(func(key, value any) bool {
		l, ok := value.(*internalListener[T])
		if ok && l != nil {
			l.dataChan <- data
		}
		return true
	})

	return nil
}

// Accepts optional buffer size for channel, if multiple buffers are provided
// , final buffer will be sum of all provided buffers
func NewBroadcastChannel[T any](buffer ...int) BroadcastInterface[T] {
	b := 0
	for _, buf := range buffer {
		b += buf
	}

	return &broadcastChannel[T]{
		buffer:    b,
		chans:     sync.Map{},
		isClosed:  false,
		closeLock: sync.Mutex{},
		singleOp:  &sync.Once{},
	}
}
