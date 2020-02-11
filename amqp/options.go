package amqp

type OptionsFn func(*Options)
type Options struct {
	*ConnectionOptions
	*ChannelOptions
}
