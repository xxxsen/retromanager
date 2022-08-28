package importer

import (
	"context"
	"fmt"
	"log"
	"path"
	"retromanager/client"
	"retromanager/cmd/tools/import_from_dir/gamelist"
	"retromanager/proto/retromanager/gameinfo"
	"time"

	"github.com/xxxsen/errs"
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

func (p *Importer) Validate() error {
	if err := p.gl.Validate(); err != nil {
		return fmt.Errorf("game validate fail, err:%w", err)
	}
	return nil
}

func (p *Importer) DoImport(ctx context.Context) error {
	for _, item := range p.gl.Games {
		if err := p.importOneGame(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (p *Importer) uploadImage(ctx context.Context, game *gamelist.GameItem) ([]string, error) {
	images := game.GetImages()
	rs := make([]string, len(images))
	for _, item := range images {
		req := &client.UploadImageRequest{
			File: p.gl.BuildFullPath(item),
		}
		rsp, err := p.client.UploadImage(ctx, req)
		if err != nil {
			return nil, errs.Wrap(errs.ErrIO, "upload image fail", err)
		}
		rs = append(rs, rsp.Meta.DownKey)
	}
	return rs, nil
}

func (p *Importer) uploadVideo(ctx context.Context, game *gamelist.GameItem) ([]string, error) {
	videos := game.GetVideos()
	rs := make([]string, len(videos))
	for _, item := range videos {
		req := &client.UploadVideoRequest{
			File: p.gl.BuildFullPath(item),
		}
		rsp, err := p.client.UploadVideo(ctx, req)
		if err != nil {
			return nil, errs.Wrap(errs.ErrIO, "upload video fail", err)
		}
		rs = append(rs, rsp.Meta.DownKey)
	}
	return rs, nil
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

func (p *Importer) uploadGameData(ctx context.Context, game *gamelist.GameItem) ([]string, []string, *client.FileMeta, error) {
	run := runner.New(10)
	var imagelist []string
	var videolist []string
	var romMeta *client.FileMeta

	run.Add("upload_image", func(ctx context.Context) error {
		var err error
		imagelist, err = p.uploadImage(ctx, game)
		return err
	}).Add("upload_video", func(ctx context.Context) error {
		var err error
		videolist, err = p.uploadVideo(ctx, game)
		return err
	}).Add("upload_rom", func(ctx context.Context) error {
		var err error
		romMeta, err = p.uploadRom(ctx, game)
		return err
	})
	if err := run.Run(ctx); err != nil {
		return nil, nil, nil, errs.Wrap(errs.ErrIO, "upload fail", err)
	}
	return imagelist, videolist, romMeta, nil
}

func (p *Importer) importOneGame(ctx context.Context, game *gamelist.GameItem) error {
	imagelist, videolist, rominfo, err := p.uploadGameData(ctx, game)
	if err != nil {
		return err
	}
	now := uint64(time.Now().UnixMilli())
	req := &client.CreateGameRequest{
		Item: &gameinfo.GameInfo{
			Platform:    proto.Uint32(uint32(p.c.system)),
			DisplayName: proto.String(game.Name),
			FileSize:    proto.Uint64(uint64(rominfo.Size)),
			Desc:        nil,
			CreateTime:  proto.Uint64(now),
			UpdateTime:  proto.Uint64(now),
			Hash:        proto.String(rominfo.MD5),
			Extinfo: &gameinfo.GameExtInfo{
				Genre:       []string{},
				Video:       videolist,
				Image:       imagelist,
				Rating:      proto.Float64(game.Rating),
				Developer:   proto.String(game.Developer),
				Publisher:   proto.String(game.Publisher),
				Releasedate: proto.String(game.ReleaseDate),
				Players:     proto.Uint32(uint32(game.GetMaxPlayerCount())),
			},
			DownKey:  proto.String(rominfo.DownKey),
			FileName: proto.String(path.Base(game.Path)),
		},
	}
	rsp, err := p.client.CreateGame(ctx, req)
	if err != nil {
		return errs.Wrap(errs.ErrServiceInternal, "create game fail", err)
	}
	log.Printf("create game succ, gameid:%d, system:%d, name:%s", rsp.GetGameId(), p.c.system, game.Name)
	return nil
}
