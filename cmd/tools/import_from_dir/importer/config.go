package importer

type config struct {
	system  int
	dir     string
	apisvr  string
	filesvr string
	ak, sk  string
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

func WithFileSvr(svr string) Option {
	return func(c *config) {
		c.filesvr = svr
	}
}

func WithSecret(ak, sk string) Option {
	return func(c *config) {
		c.ak, c.sk = ak, sk
	}
}

func WithDir(dir string) Option {
	return func(c *config) {
		c.dir = dir
	}
}
