package treqs

type Options struct {
	Verbose bool
}

type Option func(*Options)

func WithVerboseLogging(opts *Options) {
	opts.Verbose = true
}
