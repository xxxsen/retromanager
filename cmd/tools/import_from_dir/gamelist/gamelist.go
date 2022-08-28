package gamelist

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/xxxsen/errs"
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

func (gl *GameList) Validate() error {
	for _, item := range gl.Games {
		checkLst := []string{
			item.Path,
			item.Boxart,
			item.Image,
			item.Marquee,
			item.Video,
			item.Screenshot,
			item.Screentitle,
		}
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

func (item *GameItem) GetImages() []string {
	rs := make([]string, 0, 5)
	if len(item.Image) > 0 {
		rs = append(rs, item.Image)
	}
	if len(item.Boxart) > 0 {
		rs = append(rs, item.Boxart)
	}
	if len(item.Marquee) > 0 {
		rs = append(rs, item.Marquee)
	}
	if len(item.Screenshot) > 0 {
		rs = append(rs, item.Screenshot)
	}
	if len(item.Screentitle) > 0 {
		rs = append(rs, item.Screentitle)
	}
	return rs
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
