package gamelist

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
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
}

func ParseData(raw []byte) (*GameList, error) {
	gl := &GameList{}
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.Strict = false
	if err := dec.Decode(gl); err != nil {
		return nil, errs.Wrap(errs.ErrUnmarshal, "decode xml fail", err)
	}
	return gl, nil
}

func Parse(file string) (*GameList, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "read file fail", err)
	}
	return ParseData(raw)
}

func isFileExist(path, sub string) (bool, error) {
	if len(sub) == 0 {
		return true, nil
	}
	sub = strings.TrimPrefix(sub, "./")
	full := path + "/" + sub
	_, err := os.Stat(full)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func Validate(path string, gl *GameList) error {
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
			exist, err := isFileExist(path, ch)
			if err != nil {
				return fmt.Errorf("path:%s, sub:%s check err:%v", path, ch, err)
			}
			if !exist {
				return fmt.Errorf("path:%s, sub:%s not exist", path, ch)
			}
		}
	}
	return nil
}
