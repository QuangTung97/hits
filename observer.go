package hits

import (
	"hits/rpc"
	"sync"
)

const observerChannelSize = 1024

type observerService struct {
	mut        sync.Mutex
	idSequence uint32
	channels   map[uint32]chan MarshalledEvent
}

func newObserverService() *observerService {
	return &observerService{
		idSequence: 0,
	}
}

func (s *observerService) Listen(
	req *rpc.ListenRequest, listenEvents rpc.ObserverService_ListenServer,
) error {
	channel := make(chan MarshalledEvent, observerChannelSize)

	s.mut.Lock()
	id := s.idSequence
	s.idSequence++
	s.channels[id] = channel
	s.mut.Unlock()

	defer func() {
		s.mut.Lock()
		delete(s.channels, id)
		s.mut.Unlock()
	}()

	for {
		event := <-channel
		err := listenEvents.Send(&rpc.ListenEvent{
			Type:      uint32(event.Type),
			Sequence:  event.Sequence,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		})
		if err != nil {
			return err
		}
	}
}

func (s *observerService) getChannels(result *[]chan<- MarshalledEvent) {
	s.mut.Lock()
	for _, value := range s.channels {
		*result = append(*result, value)
	}
	s.mut.Unlock()
}
