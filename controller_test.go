package broadcastchannels_test

import (
	broadcastchannels "broadcast-channels"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*

Testcases need to be written for following cases :
1. Broadcasting works : case when rate of broadcasting is more than rate of receiving.
2. Unsubscription works : case when id does not exist, and when id exists
3. Subscription Works
4. closing broadcast works : case when listeners are present

*/

func TestThatRateOfBroadcastingIsMoreThanRateOfReceivingAndChannelIsClosedInBetween(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()
	id2, _ := bc.Subscribe()

	result := make([]int, 0)

	go func() {
		l, err := bc.Listener(id2)
		if l == nil || err != nil {
			return
		}

		for v := range l {
			fmt.Println(v)
			time.Sleep(2 * time.Second)
			result = append(result, v)
		}

		fmt.Println("I am done")
	}()

	for i := range 10 {
		go bc.Broadcast(i)
	}

	time.Sleep(3 * time.Second)
	bc.Close()
	assert.NotEmpty(t, result, "array should not be empty")

}

func TestThatUnsubscriptionWorksWhenNotFound(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()

	assert.NotPanics(t, func() { bc.Unsubscribe("nonexistentid1") }, "code should not panic")
}

func TestThatMultipleGoroutinesRecieveMessage(t *testing.T) {

	result := make([]int, 0, 10)

	bc := broadcastchannels.NewBroadcastChannel[int]()
	for range 10 {
		id, _ := bc.Subscribe()
		go func() {
			ch, err := bc.Listener(id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			k, ok := <-ch
			if !ok {
				fmt.Println("closed channel")
				return
			}
			result = append(result, k)
			fmt.Println("i am done", id)
		}()
	}

	bc.Broadcast(7)

	time.Sleep(2 * time.Second)

	assert.ElementsMatch(t, result, []int{7, 7, 7, 7, 7, 7, 7, 7, 7, 7}, "elements should be equal")
}

func TestThatProcessCanUnsubscribeFromBroadcastChannel(t *testing.T) {

	bc := broadcastchannels.NewBroadcastChannel[int]()
	result := make([]int, 0, 2)

	id1, _ := bc.Subscribe()
	id2, _ := bc.Subscribe()

	go func() {
		ch, err := bc.Listener(id1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		vm, ok := <-ch
		fmt.Println(vm, ok, id1)

		result = append(result, 7)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("started second")
		l, err := bc.Listener(id2)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if l == nil {
			fmt.Println("nil channel")
		}
		v, ok := <-l
		if !ok {
			fmt.Println("closed channel")
		}
		fmt.Println(v, id2)
		result = append(result, 7)
	}()
	// time.Sleep(1 * time.Second)

	bc.Unsubscribe(id2)

	time.Sleep(1 * time.Second)

	bc.Broadcast(7)

	time.Sleep(4 * time.Second)

	assert.ElementsMatch(t, result, []int{7}, "elements should be equal")

}

func TestThatClosingBroadcastIntimatesListeners(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()

	id1, _ := bc.Subscribe()
	id2, _ := bc.Subscribe()

	go func() {
		l, _ := bc.Listener(id1)
		if l == nil {
			return
		}

		for v := range l {
			fmt.Println(v)
		}
		fmt.Println("ended", id1)

	}()

	go func() {
		l, _ := bc.Listener(id2)
		if l == nil {
			return
		}

		for v := range l {
			fmt.Println(v)
		}

		fmt.Println("ended", id2)
	}()

	time.Sleep(1 * time.Second)
	bc.Close()

	time.Sleep(5 * time.Second)

}
