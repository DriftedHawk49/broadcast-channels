package broadcastchannels

import (
	"broadcast-channels/constants"
	identitylocks "broadcast-channels/identityLocks"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

/*
More functions for support in future :
1. Reset : Removes all listeners, probably for fresh re-use
*/
type BroadcastInterface[T any] interface {
	/*Id is returned for getting listeners, */
	Subscribe() (string, error)
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
	/*
		Get Listener by id. id is generated at the time of subscription.
		in case id does not exist, or Close has been called on broadcast channel, this will return a nil channel.
		Note : checking for nil-ness is important, as it will block forever if read from, unchecked.
	*/
	Listener(id string) (<-chan T, error)
}

type broadcastChannel[T any] struct {
	chans     sync.Map                    // map that will contain all the channels who have subscribed to this channel
	buffer    int                         // buffer for channels
	isClosed  bool                        // tracks the request of closing of broadcast channel
	closeLock *identitylocks.IdentityLock // lock for operations
	singleOp  sync.Once                   // for closing broadcast channel only once even if multiple process call it
}

// Use Id returned by this function to get particular listener
func (b *broadcastChannel[T]) Subscribe() (string, error) {

	if b.isClosed {
		return "", errors.New(constants.MessageClosedChannel)
	}

	if locked, id := b.closeLock.IsLocked(); locked && id == constants.CloserId {
		// close lock is held. that means, channel is being closed. We cannot subscribe to a closing channel
		// lock should be held by closer
		return "", errors.New(constants.MessageClosedChannel)
	}

	id := uuid.New().String()
	b.chans.Store(id, make(chan T, b.buffer))
	return id, nil
}

// Closes the channel
func (b *broadcastChannel[T]) Close() {

	if b.isClosed {
		return
	}
	b.isClosed = true

	// should be run only once even if race occurs to close the channel
	b.singleOp.Do(func() {
		b.closeLock.Lock(constants.CloserId)
		defer b.closeLock.Unlock()
		b.chans.Range(func(key, value any) bool {
			c := value.(chan T)
			if c != nil {
				close(c)
			}
			return true
		})
	})

}

func (b *broadcastChannel[T]) Unsubscribe(id string) {
	v, ok := b.chans.LoadAndDelete(id)
	if !ok {
		return
	}

	close(v.(chan T))
}

func (b *broadcastChannel[T]) Broadcast(data T) error {

	if b.isClosed {
		return errors.New(constants.MessageClosedChannel)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	b.chans.Range(func(key, value any) bool {
		c, ok := value.(chan T)
		if ok {
			c <- data
		}
		return true
	})

	return nil
}

// If there is no subscription with provided id, then it will return nil, hence nil check is necessary
func (b *broadcastChannel[T]) Listener(id string) (<-chan T, error) {

	if b.isClosed {
		return nil, errors.New("closed channel")
	}

	v, ok := b.chans.Load(id)
	if !ok {
		return nil, errors.New("channel not found")
	}
	return v.(chan T), nil
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
		closeLock: identitylocks.NewIdentityLock(),
		singleOp:  sync.Once{},
	}
}
