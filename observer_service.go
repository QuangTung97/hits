package hits

import (
	"context"
	"github.com/QuangTung97/hits/rpc"
	"sync"

	"google.golang.org/grpc"
)

const observerChannelSize = 1024

type observerService struct {
	mut            sync.Mutex
	idSequence     uint32
	channels       map[uint32]chan MarshalledEvent
	changeSequence uint64
	lastEvent      NullMarshalledEvent
}

func newObserverService() *observerService {
	return &observerService{
		idSequence:     0,
		channels:       make(map[uint32]chan MarshalledEvent),
		changeSequence: 0,
		lastEvent: NullMarshalledEvent{
			Valid: false,
		},
	}
}

func marshalledEventToProto(event MarshalledEvent) *rpc.ListenEvent {
	return &rpc.ListenEvent{
		Type:      uint32(event.Type),
		Sequence:  event.Sequence,
		Timestamp: event.Timestamp,
		Data:      event.Data,
	}
}

func (s *observerService) Listen(
	req *rpc.ListenRequest, listenEvents rpc.ObserverService_ListenServer,
) error {
	channel := make(chan MarshalledEvent, observerChannelSize)

	s.mut.Lock()
	id := s.idSequence
	s.channels[id] = channel
	s.idSequence++
	s.changeSequence++
	lastEvent := s.lastEvent
	s.mut.Unlock()

	defer func() {
		s.mut.Lock()
		delete(s.channels, id)
		s.changeSequence++
		s.mut.Unlock()
	}()

	if lastEvent.Valid {
		err := listenEvents.Send(marshalledEventToProto(lastEvent.Event))
		if err != nil {
			return err
		}
	}

	for {
		event := <-channel
		err := listenEvents.Send(marshalledEventToProto(event))
		if err != nil {
			return err
		}
	}
}

func (s *observerService) getChannels(changeSequence uint64,
) (uint64, []chan<- MarshalledEvent) {
	result := make([]chan<- MarshalledEvent, 0, len(s.channels))

	s.mut.Lock()
	defer s.mut.Unlock()

	if s.changeSequence == changeSequence {
		return s.changeSequence, result
	}
	for _, ch := range s.channels {
		result = append(result, ch)
	}
	return s.changeSequence, result
}

func (s *observerService) setLastEvent(lastEvent MarshalledEvent) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.lastEvent = NullMarshalledEvent{
		Valid: true,
		Event: lastEvent,
	}
}

// Listen is a client API
func Listen(ctx context.Context, address string, ch chan<- MarshalledEvent) error {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := rpc.NewObserverServiceClient(conn)
	events, err := client.Listen(ctx, &rpc.ListenRequest{})
	if err != nil {
		return err
	}
	for {
		event, err := events.Recv()
		if err != nil {
			return err
		}
		ch <- MarshalledEvent{
			Type:      EventType(event.Type),
			Sequence:  event.Sequence,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		}
	}
}
