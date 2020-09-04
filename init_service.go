package hits

import (
	"context"

	"github.com/QuangTung97/hits/rpc"

	"google.golang.org/grpc"
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
	req *rpc.ReadEventsRequest, outEvents rpc.InitService_ReadEventsServer,
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

func ReadEventsFrom(ctx context.Context, address string,
	initSequence uint64, callback func(event MarshalledEvent),
) error {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	client := rpc.NewInitServiceClient(conn)
	req := &rpc.ReadEventsRequest{
		FromSequence: initSequence,
	}
	events, err := client.ReadEvents(ctx, req)
	if err != nil {
		return err
	}
	for {
		event, err := events.Recv()
		if err != nil {
			return err
		}

		callback(MarshalledEvent{
			Type:      EventType(event.Type),
			Sequence:  event.Sequence,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		})
	}
}
