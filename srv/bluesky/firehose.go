package blueskysrv

import (
	cbor2 "github.com/fxamacker/cbor/v2"
	atprotorecord "github.com/reiver/go-atproto/record"
	atprotosync "github.com/reiver/go-atproto/com/atproto/sync"
	_ "github.com/reiver/go-bsky"
	"github.com/reiver/go-erorr"
	"github.com/reiver/go-iter"
	"github.com/reiver/go-reg"

	"github.com/reiver/roodmagi/lib/errors"
	. "github.com/reiver/roodmagi/srv/log"
)

// FireHose stores the list of subscribers.
//
// To subscribe to the Bluesky firehose, call:
//
//	channelName, channel, err := blueskysrv.FireHose.Subscribe()
//	if nil != err {
//		return err
//	}
//	blueskysrv.FireHose.Unsubscribe(channelName)
var FireHose RepoSubscriber

func init() {
	Log("[bluesky-service FireHose] about to spawn toiler ...")
			
//@TODO: make it so this can recover from a panic(), such as from not being able to connect to the Bluesky server.
			
	go func() {
		Log("[bluesky-service FireHose] toiler spawned!")

		FireHose.Toil()
	}()
}

type RepoSubscriber struct {
	registry reg.Registry[chan atprotorecord.Record]
}

func (receiver *RepoSubscriber) HasSubscribers() bool {
	if nil == receiver {
		return false
	}

	return 1 <= receiver.registry.Len()
}

func (receiver *RepoSubscriber) Subscribe() (string, <-chan atprotorecord.Record, error) {
	var noName string
	var noChannel chan atprotorecord.Record

	if nil == receiver {
		return noName, noChannel, errors.ErrNilReceiver
	}

	var name string = chronorand()
	var channel chan atprotorecord.Record = make(chan atprotorecord.Record, 1024)

	previous, found := receiver.registry.Set(name, channel)
	if found {
		receiver.registry.Set(name, previous)
		close(channel)
		return noName, noChannel, errFound
	}

	return name, channel, nil
}

func (receiver *RepoSubscriber) Unsubscribe(name string) error {
	if nil == receiver {
		return errors.ErrNilReceiver
	}

	receiver.registry.Unset(name)
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

//			var stuff map[any]any = map[any]any{}
			var stuff map[string]any = map[string]any{}

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

			var recordType string
			{
				var err error

				recordType, err = atprotorecord.InferType(stuff)
				if nil != err {
					Logf("[bluesky-service FireHose] iter[%d]: NOTE: error trying to get $type (so skipping): %s", iterationNumber, err)
					return
				}
			}
			Logf("[bluesky-service FireHose] iter[%d]: $type = %q", iterationNumber, recordType)

			var record atprotorecord.Record
			{
				var created bool

				record, created = atprotorecord.New(recordType)
				if !created {
					Logf("[bluesky-service FireHose] iter[%d]: NOTE: could not create (struct) record of $type = %q (so skipping)", iterationNumber, recordType)
					return
				}

				err := record.FromMap(stuff)
				if nil != err {
					Logf("[bluesky-service FireHose] iter[%d]: get error when trying to load (struct) record from map (so skipping): %s", iterationNumber, err)
					return
				}
			}
			Logf("[bluesky-service FireHose] iter[%d]: record-type = %T", iterationNumber, record)

			{
				go func(iterationNumber uint64) {
					err := receiver.publish(record)
					if nil != err {
						Logf("[bluesky-service FireHose] iter[%d]: problem publishing: %s", iterationNumber, err)
						return
					}
					Logf("[bluesky-service FireHose] iter[%d]: published!", iterationNumber)
				}(iterationNumber)
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

func (receiver *RepoSubscriber) publish(record atprotorecord.Record) error {
	if nil == receiver {
		return errors.ErrNilReceiver
	}

	receiver.registry.For(func(name string, channel chan atprotorecord.Record){
		go func() {
			Logf("[bluesky-service FireHose] sending record %q to channel %q", record.RecordType(), name)
			channel <- record
		}()
	})

	return nil
}
