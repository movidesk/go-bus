package amqp

import "errors"

var (
	ConnectionError = errors.New("Unable to connect on amqp broker")
)
