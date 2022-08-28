package importer

type config struct {
	system int
	dir    string
	apisvr string
}

type Option func(c *config)

func WithSystem(system int) Option {
	return func(c *config) {
		c.system = system
	}
}

func WithAPISvr(svr string) Option {
	return func(c *config) {
		c.apisvr = svr
	}
}

func WithDir(dir string) Option {
	return func(c *config) {
		c.dir = dir
	}
}
