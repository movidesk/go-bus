package amqp

import "time"

type OptionsFn func(*Options)
type Options struct {
	Dsn string

	ReconectInterval time.Duration
}

func SetDsn(dsn string) OptionsFn {
	return func(o *Options) {
		o.Dsn = dsn
	}
}

func SetReconectInterval(interval time.Duration) OptionsFn {
	return func(o *Options) {
		o.ReconectInterval = interval
	}
}
