package gamelist

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/xxxsen/common/errs"
)

type GameItem struct {
	XMLName     xml.Name `xml:"game"`
	Path        string   `xml:"path"`
	Name        string   `xml:"name"`
	Image       string   `xml:"image"`
	Boxart      string   `xml:"boxart"`
	Screenshot  string   `xml:"screenshot"`
	Screentitle string   `xml:"screentitle"`
	Marquee     string   `xml:"marquee"`
	Video       string   `xml:"video"`
	ReleaseDate string   `xml:"releasedate"`
	Developer   string   `xml:"developer"`
	Publisher   string   `xml:"publisher"`
	Lang        string   `xml:"lang"`
	Region      string   `xml:"region"`
	PlayCount   string   `xml:"playcount"`
	Gameime     int      `xml:"gametime"`
	LastPlayed  string   `xml:"lastplayed"`
	Rating      float64  `xml:"rating"`
	Desc        string   `xml:"desc"`
}

type FolderItem struct {
	XMLName  xml.Name `xml:"folder"`
	Path     string   `xml:"path"`
	Name     string   `xml:"name"`
	Sortname string   `xml:"sortname"`
	Image    string   `xml:"image"`
}

type GameList struct {
	XMLName xml.Name      `xml:"gameList"`
	Games   []*GameItem   `xml:"game"`
	Folders []*FolderItem `xml:"folder"`
	root    string        `xml:"-"`
}

func New() *GameList {
	return &GameList{}
}

func (gl *GameList) parseData(raw []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.Strict = false
	if err := dec.Decode(gl); err != nil {
		return errs.Wrap(errs.ErrUnmarshal, "decode xml fail", err)
	}
	return nil
}

func (gl *GameList) FolderSize() int {
	return len(gl.Folders)
}

func (gl *GameList) GameSize() int {
	return len(gl.Games)
}

//Clean 重建游戏列表, 将列表中异常的数据剔除掉
func (gl *GameList) Clean() error {
	games := make([]*GameItem, 0, len(gl.Games))
	for _, game := range gl.Games {
		if exist, _ := gl.isFileExist(game.Path); !exist {
			continue
		}
		if exist, _ := gl.isFileExist(game.Image); !exist {
			game.Image = ""
		}
		if exist, _ := gl.isFileExist(game.Boxart); !exist {
			game.Boxart = ""
		}
		if exist, _ := gl.isFileExist(game.Marquee); !exist {
			game.Marquee = ""
		}
		if exist, _ := gl.isFileExist(game.Screenshot); !exist {
			game.Screenshot = ""
		}
		if exist, _ := gl.isFileExist(game.Screentitle); !exist {
			game.Screentitle = ""
		}
		if exist, _ := gl.isFileExist(game.Video); !exist {
			game.Video = ""
		}
		games = append(games, game)
	}
	gl.Games = games
	return nil
}

func (gl *GameList) Parse(file string) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return errs.Wrap(errs.ErrIO, "read file fail", err)
	}
	if err := gl.parseData(raw); err != nil {
		return err
	}
	gl.root = path.Dir(file)
	if !strings.HasSuffix(gl.root, "/") {
		gl.root += "/"
	}
	return nil
}

func (gl *GameList) isFileExist(sub string) (bool, error) {
	if len(sub) == 0 {
		return true, nil
	}
	full := gl.BuildFullPath(sub)
	_, err := os.Stat(full)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func (item *GameItem) fileList() []string {
	return []string{
		item.Path,
		item.Boxart,
		item.Image,
		item.Marquee,
		item.Video,
		item.Screenshot,
		item.Screentitle,
	}
}

func (gl *GameList) Validate() error {
	for _, item := range gl.Games {
		checkLst := item.fileList()
		for _, ch := range checkLst {
			exist, err := gl.isFileExist(ch)
			if err != nil {
				return fmt.Errorf("sub:%s check err:%v", ch, err)
			}
			if !exist {
				return fmt.Errorf("sub:%s not exist", ch)
			}
		}
	}
	return nil
}

func (gl *GameList) RemovePrefix(loc string) string {
	return strings.TrimLeft(loc, "./")
}

func (gl *GameList) BuildFullPath(loc string) string {
	return gl.root + gl.RemovePrefix(loc)
}

func (item *GameItem) GetImage() string {
	return item.Image
}

func (item *GameItem) GetMaxPlayerCount() int {
	if len(item.PlayCount) == 0 {
		return 1
	}
	arr := strings.Split(item.PlayCount, "-")
	val, _ := strconv.ParseUint(arr[len(arr)-1], 0, 64)
	if val == 0 {
		val = 1
	}
	return int(val)
}

func (item *GameItem) GetVideos() []string {
	if len(item.Video) == 0 {
		return nil
	}
	return []string{
		item.Video,
	}
}

func (item *GameItem) GetRom() string {
	return item.Path
}
