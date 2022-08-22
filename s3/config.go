package s3

type config struct {
	secretId  string
	secretKey string
	ssl       bool
	endpoint  string
	bucket    string
}

type Option func(c *config)

func WithSecret(id, key string) Option {
	return func(c *config) {
		c.secretId = id
		c.secretKey = key
	}
}

func WithSSL(v bool) Option {
	return func(c *config) {
		c.ssl = v
	}
}

func WithEndpoint(ep string) Option {
	return func(c *config) {
		c.endpoint = ep
	}
}

func WithBucket(bk string) Option {
	return func(c *config) {
		c.bucket = bk
	}
}
