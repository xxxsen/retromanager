package client

type config struct {
	apisvr  string
	filesvr string
	ak      string
	sk      string
}

type Option func(c *config)

func WithAPISvr(host string) Option {
	return func(c *config) {
		c.apisvr = host
	}
}

func WithFileSvr(host string) Option {
	return func(c *config) {
		c.filesvr = host
	}
}

func WithSecret(ak, sk string) Option {
	return func(c *config) {
		c.ak, c.sk = ak, sk
	}
}
