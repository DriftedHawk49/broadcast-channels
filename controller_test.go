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

func TestThatMultipleGoroutinesRecieveMessage(t *testing.T) {

	result := make([]int, 0, 10)

	bc := broadcastchannels.NewBroadcastChannel[int]()
	for range 10 {
		id := bc.Subscribe()
		go func() {
			k := <-bc.Listener(id)
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

	id1 := bc.Subscribe()
	id2 := bc.Subscribe()

	go func() {
		vm, ok := <-bc.Listener(id1)
		fmt.Println(vm, ok, id1)

		result = append(result, 7)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("started second")
		l := bc.Listener(id2)
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
