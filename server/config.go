package server

type Config struct {
	addresses  []string
	registerFn HandlerRegisterFunc
	attach     map[string]interface{}
}

type Option func(c *Config)

func WithHandlerRegister(fn HandlerRegisterFunc) Option {
	return func(c *Config) {
		c.registerFn = fn
	}
}

func WithAddress(address string) Option {
	return func(c *Config) {
		c.addresses = append(c.addresses, address)
	}
}

func WithAttach(key string, attach interface{}) Option {
	return func(c *Config) {
		c.attach[key] = attach
	}
}
