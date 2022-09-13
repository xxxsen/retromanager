package importer

import (
	"context"
	"fmt"
	"log"
	"path"
	"retromanager/client"
	"retromanager/gamelist"
	"retromanager/proto/retromanager/gameinfo"
	"time"

	"github.com/xxxsen/common/errs"
	"github.com/xxxsen/runner"
	"google.golang.org/protobuf/proto"
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

func (p *Importer) GetGameList() *gamelist.GameList {
	return p.gl
}

func (p *Importer) Clean() error {
	if err := p.gl.Clean(); err != nil {
		return fmt.Errorf("clean game info fail, err:%w", err)
	}
	return nil
}

func (p *Importer) Validate() error {
	if err := p.gl.Validate(); err != nil {
		return fmt.Errorf("game validate fail, err:%w", err)
	}
	return nil
}

func (p *Importer) DoImport(ctx context.Context) error {
	run := runner.New(5)
	for idx, item := range p.gl.Games {
		item := item
		run.Add(fmt.Sprintf("import_%d", idx), func(ctx context.Context) error {
			if err := p.importOneGame(ctx, item); err != nil {
				return err
			}
			return nil
		})

	}
	return run.Run(ctx)
}

func (p *Importer) uploadImage(ctx context.Context, path string) (*string, error) {
	if len(path) == 0 {
		return nil, nil
	}
	req := &client.UploadImageRequest{
		File: p.gl.BuildFullPath(path),
	}
	rsp, err := p.client.UploadImage(ctx, req)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "upload image fail", err)
	}
	return &rsp.Meta.DownKey, nil
}

func (p *Importer) uploadVideo(ctx context.Context, path string) (*string, error) {
	if len(path) == 0 {
		return nil, nil
	}
	req := &client.UploadVideoRequest{
		File: p.gl.BuildFullPath(path),
	}
	rsp, err := p.client.UploadVideo(ctx, req)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "upload video fail", err)
	}
	return &rsp.Meta.DownKey, nil
}

func (p *Importer) uploadRom(ctx context.Context, game *gamelist.GameItem) (*client.FileMeta, error) {
	req := &client.UploadFileRequest{
		File: p.gl.BuildFullPath(game.Path),
	}
	rsp, err := p.client.UploadFile(ctx, req)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "upload rom fail", err)
	}
	return rsp.Meta, nil
}

type uploadContext struct {
	image       *string
	video       *string
	romMeta     *client.FileMeta
	marquee     *string
	boxart      *string
	screentitle *string
	screenshot  *string
}

func (p *Importer) uploadGameData(ctx context.Context, game *gamelist.GameItem) (*uploadContext, error) {
	uctx := &uploadContext{}

	run := runner.New(10)
	run.Add("upload_image", func(ctx context.Context) error {
		var err error
		uctx.image, err = p.uploadImage(ctx, game.Image)
		return err
	}).Add("upload_video", func(ctx context.Context) error {
		var err error
		uctx.video, err = p.uploadVideo(ctx, game.Video)
		return err
	}).Add("upload_rom", func(ctx context.Context) error {
		var err error
		uctx.romMeta, err = p.uploadRom(ctx, game)
		return err
	}).Add("upload_marquee", func(ctx context.Context) error {
		var err error
		uctx.marquee, err = p.uploadImage(ctx, game.Marquee)
		return err
	}).Add("upload_boxart", func(ctx context.Context) error {
		var err error
		uctx.boxart, err = p.uploadImage(ctx, game.Boxart)
		return err
	}).Add("upload_screentitle", func(ctx context.Context) error {
		var err error
		uctx.screentitle, err = p.uploadImage(ctx, game.Screenshot)
		return err
	}).Add("upload_screenshot", func(ctx context.Context) error {
		var err error
		uctx.screenshot, err = p.uploadImage(ctx, game.Screenshot)
		return err
	})
	if err := run.Run(ctx); err != nil {
		return nil, errs.Wrap(errs.ErrIO, "upload fail", err)
	}
	return uctx, nil
}

func (p *Importer) importOneGame(ctx context.Context, game *gamelist.GameItem) error {
	uctx, err := p.uploadGameData(ctx, game)
	if err != nil {
		return err
	}
	now := uint64(time.Now().UnixMilli())
	desc := game.Desc
	if len(desc) == 0 {
		desc = "default"
	}
	item := &gameinfo.GameInfo{
		Platform:    proto.Uint32(uint32(p.c.system)),
		DisplayName: proto.String(game.Name),
		FileSize:    proto.Uint64(uint64(uctx.romMeta.Size)),
		Desc:        proto.String(desc),
		CreateTime:  proto.Uint64(now),
		UpdateTime:  proto.Uint64(now),
		Hash:        proto.String(uctx.romMeta.MD5),
		Extinfo: &gameinfo.GameExtInfo{
			Genre:       []string{},
			Rating:      proto.Float64(game.Rating),
			Developer:   proto.String(game.Developer),
			Publisher:   proto.String(game.Publisher),
			Releasedate: proto.String(game.ReleaseDate),
			Players:     proto.Uint32(uint32(game.GetMaxPlayerCount())),
			Lang:        proto.String(game.Lang),
			Region:      proto.String(game.Region),
			Marquee:     uctx.marquee,
			Boxart:      uctx.boxart,
			Screenshot:  uctx.screenshot,
			Screentitle: uctx.screentitle,
		},
		DownKey:  proto.String(uctx.romMeta.DownKey),
		FileName: proto.String(path.Base(game.Path)),
	}
	if uctx.image != nil && len(*uctx.image) > 0 {
		item.Extinfo.Image = []string{*uctx.image}
	}
	if uctx.video != nil && len(*uctx.video) > 0 {
		item.Extinfo.Image = []string{*uctx.video}
	}
	req := &client.CreateGameRequest{
		Item: item,
	}
	rsp, err := p.client.CreateGame(ctx, req)
	if err != nil {
		return errs.Wrap(errs.ErrServiceInternal, "create game fail", err)
	}
	log.Printf("create game succ, gameid:%d, system:%d, name:%s", rsp.GetGameId(), p.c.system, game.Name)
	return nil
}
