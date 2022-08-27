package gamelist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<gameList xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<folder>
			<path>./chinese</path>
			<name>GBA中文游戏大全</name>
			<sortname>01 =- GBA游戏大全</sortname>
			<image>./media/chinese.png</image>
		</folder>
		<folder>
			<path>./top100</path>
			<name>MC百强GBA游戏</name>
			<sortname>02 =- MC百强GBA游戏</sortname>
			<image>./media/top100.png</image>
		</folder>
		<game>
			<path>./pokemon/口袋妖怪噩の幻V.1.0.zip</path>
			<name>口袋妖怪噩の幻V.1.0</name>
			<image>./previews/口袋妖怪噩の幻V.1.0.png</image>
		</game>
		<game>
			<path>./pokemon/口袋妖怪大乱斗完整版.zip</path>
			<name>口袋妖怪大乱斗完整版</name>
			<image>./previews/口袋妖怪大乱斗完整版.png</image>
		</game>
		<game>
			<path>./pokemon/口袋妖怪火红超梦反击.zip</path>
			<name>口袋妖怪火红超梦反击</name>
			<image>./previews/口袋妖怪火红超梦反击.png</image>
		</game>
	</gameList>`
	gl, err := ParseData([]byte(data))
	assert.NoError(t, err)
	t.Logf("%+v", *gl)
}
