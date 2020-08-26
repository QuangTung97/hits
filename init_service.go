package hits

import (
	"hits/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type initService struct {
	ctx *Context
}

func newInitService(ctx *Context) *initService {
	return &initService{
		ctx: ctx,
	}
}

func (s *initService) ReadEvents(
	req *rpc.ReadEventsRequest, outEvents rpc.ObserverService_ListenServer,
) error {
	events, err := s.ctx.callbacks.journaler.ReadFrom(req.FromSequence)
	if err == ErrEventsNotFound {
		st := status.New(codes.FailedPrecondition, "events from this sequence not found")
		st, err := st.WithDetails(&rpc.ErrEventsNotFound{})
		if err != nil {
			return err
		}
		return st.Err()
	}
	if err != nil {
		return err
	}

	for _, e := range events {
		err := outEvents.Send(marshalledEventToProto(e))
		if err != nil {
			return err
		}
	}

	return nil
}
