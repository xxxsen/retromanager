package importer

import (
	"context"
	"fmt"
	"retromanager/client"
	"retromanager/cmd/tools/import_from_dir/gamelist"
)

type Importer struct {
	c      *config
	client *client.Client
	gl     *gamelist.GameList
}

func New(opts ...Option) (*Importer, error) {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	cli, err := client.New(client.WithHost(c.apisvr))
	if err != nil {
		return nil, fmt.Errorf("create api client fail, err:%w", err)
	}
	gl := gamelist.New()
	gamelistFile := fmt.Sprintf("%s/gamelist.xml", c.dir)
	if err := gl.Parse(gamelistFile); err != nil {
		return nil, fmt.Errorf("parse gamelist fail, err:%w", err)
	}
	return &Importer{c: c, client: cli, gl: gl}, nil
}

func (p *Importer) Validate() error {
	if err := p.gl.Validate(); err != nil {
		return fmt.Errorf("game validate fail, err:%w", err)
	}
	return nil
}

func (p *Importer) DoImport(ctx context.Context) error {
	//TODO: finish it
	panic(1)
}
