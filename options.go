package hits

import "google.golang.org/grpc"

type runOptions struct {
	onServer func(server *grpc.Server)
}

type RunOption func(opts *runOptions)

func newRunOptions() *runOptions {
	return &runOptions{}
}

func WithOnServer(fn func(server *grpc.Server)) RunOption {
	return func(opts *runOptions) {
		opts.onServer = fn
	}
}

func (r *runOptions) apply(opts ...RunOption) {
	for _, opt := range opts {
		opt(r)
	}
}
