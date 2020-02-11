package amqp

import "errors"

var (
	ConnectionError = errors.New("Unable to connect to amqp broker")
)
