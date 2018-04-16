package mqtt

type Config struct {
	KeepAlive int `yaml:"keepAlive"`
	// The number of seconds to wait for the CONNECT message before disconnecting.
	// If not set then default to 2 seconds.
	ConnectTimeout int `yaml:"connectTimeout"`
	// The number of seconds to wait for any ACK messages before failing.
	// If not set then default to 20 seconds.
	AckTimeout int `yaml:"ackTimeout"`
	// The number of times to retry sending a packet if ACK is not received.
	// If no set then default to 3 retries.
	TimeoutRetries int `yaml:"timeoutRetries"`
}
