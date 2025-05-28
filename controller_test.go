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

BASIC TEST CASES
	1. Initialisation works
	2. Subscription works
	3. Unsubscribe works
	4. Closing of channel works

ADVANCED TEST CASES
	1. Subscription returns listners and id
	2. Unsubscription works without race when broadcast is going on.
	3. Closing of channel when broadcast is triggered after close command.
	4. rate of broadcast is more than rate of consumption
	5. closing of channel when doneChan is not utilised
	6. Broadcast when there are no subscriptions
	7. Closing a channel from multiple go routines
*/

// BASIC TEST CASES

func TestThatInitialisationWorks(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()

	assert.NotNil(t, bc, "broadcasting channel composer should not be nil")
}

func TestThatSubscriptionWorks(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()

	assert.NotNil(t, bc, "broadcasting channel composer should not be nil")

	id1, sub1, err := bc.Subscribe()
	assert.Nil(t, err, "error should be nil")

	assert.NotEmpty(t, id1, "returned id should not be empty")
	assert.NotNil(t, sub1, "sub listners should not be nil")

	result := make([]int, 1)

	go func() {
		for {
			select {
			case v, ok := <-sub1.DataChan:
				if ok {
					result[0] = v
				}

			case <-sub1.DoneChan:
				fmt.Println("Closing it")
				return
			}
		}
	}()

	bc.Broadcast(2)

	time.Sleep(1 * time.Second)
	bc.Close()
	time.Sleep(1 * time.Second)

	assert.ElementsMatch(t, []int{2}, result, "elements must match as a sign that subscription worked")
}

func TestThatUnsubscribeWorks(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()

	result := make([]int, 0)

	_, l1, err := bc.Subscribe()
	assert.Nil(t, err, "error should be nil")

	id2, l2, err := bc.Subscribe()
	assert.Nil(t, err, "error should be nil")

	go func() {
		for {
			select {
			case v := <-l1.DataChan:
				result = append(result, v)
			case <-l1.DoneChan:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case v := <-l2.DataChan:
				result = append(result, v)
			case <-l2.DoneChan:
				return
			}
		}
	}()

	time.Sleep(1 * time.Second)
	err = bc.Broadcast(3)
	assert.Nil(t, err, "error should be nil")
	time.Sleep(1 * time.Second)
	bc.Unsubscribe(id2)
	bc.Broadcast(10)

	time.Sleep(2 * time.Second)
	assert.ElementsMatch(t, []int{3, 3, 10}, result, "elements should match")

}

func TestThatClosingOfChannelWorks(t *testing.T) {
	bc := broadcastchannels.NewBroadcastChannel[int]()
	result := make([]int, 0)

	_, l1, err := bc.Subscribe()
	assert.Nil(t, err, "error should be nil")

	_, l2, err := bc.Subscribe()
	assert.Nil(t, err, "error should be nil")

	go func() {
		for {
			select {
			case v := <-l1.DataChan:
				fmt.Println(v)
			case <-l1.DoneChan:
				result = append(result, -1)
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case v := <-l2.DataChan:
				fmt.Println(v)
			case <-l2.DoneChan:
				result = append(result, -1)
				return
			}
		}
	}()

	time.Sleep(1 * time.Second)

	bc.Close()
	time.Sleep(2 * time.Second)

	assert.ElementsMatch(t, []int{-1, -1}, result, "elements should match")

}

// ADVANCED TEST CASES

// func TestThat

// func TestThatRateOfBroadcastingIsMoreThanRateOfReceivingAndChannelIsClosedInBetween(t *testing.T) {
// 	bc := broadcastchannels.NewBroadcastChannel[int]()
// 	id2, _ := bc.Subscribe()

// 	result := make([]int, 0)

// 	go func() {
// 		l, err := bc.Listener(id2)
// 		if l == nil || err != nil {
// 			return
// 		}

// 		for v := range l {
// 			fmt.Println(v)
// 			time.Sleep(2 * time.Second)
// 			result = append(result, v)
// 		}

// 		fmt.Println("I am done")
// 	}()

// 	for i := range 10 {
// 		go bc.Broadcast(i)
// 	}

// 	time.Sleep(3 * time.Second)
// 	bc.Close()
// 	assert.NotEmpty(t, result, "array should not be empty")

// }

// func TestThatUnsubscriptionWorksWhenNotFound(t *testing.T) {
// 	bc := broadcastchannels.NewBroadcastChannel[int]()

// 	assert.NotPanics(t, func() { bc.Unsubscribe("nonexistentid1") }, "code should not panic")
// }

// func TestThatMultipleGoroutinesRecieveMessage(t *testing.T) {

// 	result := make([]int, 0, 10)

// 	bc := broadcastchannels.NewBroadcastChannel[int]()
// 	for range 10 {
// 		id, _ := bc.Subscribe()
// 		go func() {
// 			ch, err := bc.Listener(id)
// 			if err != nil {
// 				fmt.Println(err.Error())
// 				return
// 			}

// 			k, ok := <-ch
// 			if !ok {
// 				fmt.Println("closed channel")
// 				return
// 			}
// 			result = append(result, k)
// 			fmt.Println("i am done", id)
// 		}()
// 	}

// 	bc.Broadcast(7)

// 	time.Sleep(2 * time.Second)

// 	assert.ElementsMatch(t, result, []int{7, 7, 7, 7, 7, 7, 7, 7, 7, 7}, "elements should be equal")
// }

// func TestThatProcessCanUnsubscribeFromBroadcastChannel(t *testing.T) {

// 	bc := broadcastchannels.NewBroadcastChannel[int]()
// 	result := make([]int, 0, 2)

// 	id1, _ := bc.Subscribe()
// 	id2, _ := bc.Subscribe()

// 	go func() {
// 		ch, err := bc.Listener(id1)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 			return
// 		}
// 		vm, ok := <-ch
// 		fmt.Println(vm, ok, id1)

// 		result = append(result, 7)
// 	}()

// 	go func() {
// 		time.Sleep(2 * time.Second)
// 		fmt.Println("started second")
// 		l, err := bc.Listener(id2)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 			return
// 		}
// 		if l == nil {
// 			fmt.Println("nil channel")
// 		}
// 		v, ok := <-l
// 		if !ok {
// 			fmt.Println("closed channel")
// 		}
// 		fmt.Println(v, id2)
// 		result = append(result, 7)
// 	}()
// 	// time.Sleep(1 * time.Second)

// 	bc.Unsubscribe(id2)

// 	time.Sleep(1 * time.Second)

// 	bc.Broadcast(7)

// 	time.Sleep(4 * time.Second)

// 	assert.ElementsMatch(t, result, []int{7}, "elements should be equal")

// }

// func TestThatClosingBroadcastIntimatesListeners(t *testing.T) {
// 	bc := broadcastchannels.NewBroadcastChannel[int]()

// 	id1, _ := bc.Subscribe()
// 	id2, _ := bc.Subscribe()

// 	go func() {
// 		l, _ := bc.Listener(id1)
// 		if l == nil {
// 			return
// 		}

// 		for v := range l {
// 			fmt.Println(v)
// 		}
// 		fmt.Println("ended", id1)

// 	}()

// 	go func() {
// 		l, _ := bc.Listener(id2)
// 		if l == nil {
// 			return
// 		}

// 		for v := range l {
// 			fmt.Println(v)
// 		}

// 		fmt.Println("ended", id2)
// 	}()

// 	time.Sleep(1 * time.Second)
// 	bc.Close()

// 	time.Sleep(5 * time.Second)

// }
