package client

type config struct {
	host string
}

type Option func(c *config)

func WithHost(host string) Option {
	return func(c *config) {
		c.host = host
	}
}
