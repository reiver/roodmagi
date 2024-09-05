package blueskysrv

import (
	"sync"

	cbor2 "github.com/fxamacker/cbor/v2"
	atprotosync "github.com/reiver/go-atproto/com/atproto/sync"
	"github.com/reiver/go-erorr"
	"github.com/reiver/go-iter"
//	"github.com/reiver/go-json"

	"github.com/reiver/roodmagi/lib/errors"
	. "github.com/reiver/roodmagi/srv/log"
)

var FireHose RepoSubscriber

func init() {
	Log("[[bluesky-service FireHose] about to spawn toiler ...")
			
//@TODO: make it so this can recover from a panic(), such as from not being able to connect to the Bluesky server.
			
	go func() {
		Log("[[bluesky-service FireHose] toiler spawned!")

		FireHose.Toil()
	}()
}

type RepoSubscriber struct {
	mutex sync.Mutex
	subscribers map[string]chan any
}

func (receiver *RepoSubscriber) HasSubscribers() bool {
	if nil == receiver {
		return false
	}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	return 1 <= len(receiver.subscribers)
}

func (receiver *RepoSubscriber) Subscribe() (string, <-chan any, error) {
	var noName string
	var noChannel chan any

	if nil == receiver {
		return noName, noChannel, errors.ErrNilReceiver
	}

	var name string = chronorand()

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	if nil == receiver.subscribers {
		receiver.subscribers = map[string]chan any{}
	}

	if  _, found := receiver.subscribers[name]; found {
		return noName, noChannel, errFound
	}

	var channel chan any = make(chan any, 1024)

	receiver.subscribers[name] = channel
	return name, channel, nil
}

func (receiver *RepoSubscriber) Unsubscribe(name string) error {
	if nil == receiver {
		return errors.ErrNilReceiver
	}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	if nil == receiver.subscribers {
		return errNotFound
	}

	channel, found := receiver.subscribers[name]
	if !found {
		return errNotFound
	}

	if nil != channel {
		close(channel)
	}

	delete(receiver.subscribers, name)

	return nil
}

func (receiver *RepoSubscriber) Toil() {
	if nil == receiver {
		var err error = erorr.Errorf("[bluesky-service FireHose] ERROR: %w", errors.ErrNilReceiver)
		panic(err)
	}

	iterator, err := atprotosync.SubscribeRepos()
	if nil != err {
		err = erorr.Errorf("[bluesky-service FireHose] ERROR: problem with atproto subscribe to 'subscribeRepos': %w", err)
		Log(err.Error())
		panic(err)
		return
	}
	if nil == iterator {
		var err error = erorr.Error("[bluesky-service FireHose] ERROR: problem with atproto subscribe to 'subscribeRepos': nil iterator")
		Log(err.Error())
		panic(err)
		return
	}
	defer iterator.Close()

	var iterationNumber uint64
	{
		err = iter.For{iterator}.Each(func(cborData []byte){
			iterationNumber++

			Logf("[bluesky-service FireHose] iter[%d]: next", iterationNumber)

			if !receiver.HasSubscribers() {
				Logf("[bluesky-service FireHose] iter[%d]: no subscribers — so skipping", iterationNumber)
				return
			}

			var stuff map[any]any = map[any]any{}

			err = cbor2.Unmarshal(cborData, &stuff)
			if nil != err {
				Logf("[bluesky-service FireHose] iter[%d]: ERROR: problem unmarshaling CBOR from block: %s", iterationNumber, err)
				Logf("[bluesky-service FireHose] iter[%d]: CBOR: %q", iterationNumber, string(cborData))
				return
			}

			// Skip anything without "$type".
//@TODO: could this be sped up by doing a byts.Contains() BEFORE the unmarshaling?
			if _, found := stuff["$type"]; !found {
				Logf("[bluesky-service FireHose] iter[%d]: NOTE: no $type (so skipping)", iterationNumber)
				return
			}

/*
			err = route.PublishEvent(func(ew httpsse.EventWriter)error{
					if nil == ew {
						return errNilHTTPSEEEventWriter
					}

					jsonBytes, err := json.Marshal(stuff)
					if nil != err {
						ew.WriteComment(" json-marshal error:")
						ew.WriteComment(" " + err.Error())
						ew.WriteComment("")
					}

					ew.WriteComment("raw-data:")
					ew.WriteComment(fmt.Sprintf("%q", cborData))
					ew.WriteEvent("com.atproto.sync.subscribeRepos")
					ew.WriteData(string(jsonBytes))

					return nil
			})
			if nil != err {
				Logf("[bluesky-service FireHose] ERROR: iter[%d]: problem publishing event to httpsse event: %s", iterationNumber, err)
				return
			}
*/
			Logf("[bluesky-service FireHose] iter[%d]: good!!!", iterationNumber)
		})
	}

	if nil != err {
		Logf("[bluesky-service FireHose] ERROR: problem with post-iterator 'subscribeRepos': %s", err)
		return
	}

	Logf("[bluesky-service FireHose] WTF: done after %d iterations?!?!?!?", iterationNumber)
}
