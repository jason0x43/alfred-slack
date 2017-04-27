package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path"
	"strings"

	alfred "github.com/jason0x43/go-alfred"
)

type spriteDesc struct {
	Name string `json:"short_name"`
	X    int    `json:"sheet_x"`
	Y    int    `json:"sheet_y"`
}

var spriteInfo []spriteDesc

func getEmojiFromSlack(name string) (filename string, err error) {
	name = strings.TrimSuffix(strings.TrimPrefix(name, ":"), ":")

	var emoji Emoji
	for i := range cache.Emoji {
		if cache.Emoji[i].Name == name {
			emoji = cache.Emoji[i]
			break
		}
	}

	if emoji.Name == "" {
		err = fmt.Errorf(`Unknown emoji "%s"`, name)
		return
	}

	filename = path.Join(emojiDir, emoji.Filename())
	if fileExists(filename) {
		return
	}

	return emoji.Retrieve(emojiDir)
}

func getEmojiFromSprite(name string) (filename string, err error) {
	name = strings.TrimSuffix(strings.TrimPrefix(name, ":"), ":")

	filename = path.Join(emojiDir, name+".png")
	if fileExists(filename) {
		return
	}

	if len(spriteInfo) == 0 {
		if err = alfred.LoadJSON(path.Join(workflow.WorkflowDir(), "emoji.json"), &spriteInfo); err != nil {
			return
		}
	}

	var desc spriteDesc
	for i := range spriteInfo {
		if spriteInfo[i].Name == name {
			desc = spriteInfo[i]
			break
		}
	}

	if desc.Name == "" {
		err = fmt.Errorf(`Unknown sprite name "%s"`, name)
		return
	}

	sprite, _ := os.Open(path.Join(workflow.WorkflowDir(), "sheet_apple_64_indexed_128.png"))
	defer sprite.Close()
	spriteImage, _, _ := image.Decode(sprite)

	emoji := image.NewRGBA(image.Rect(0, 0, 64, 64))
	draw.Draw(emoji, emoji.Bounds(), spriteImage, image.Point{desc.X * 64, desc.Y * 64}, draw.Src)

	toimg, _ := os.Create(filename)
	defer toimg.Close()

	err = png.Encode(toimg, emoji)
	return
}

func getAllSpriteEmoji() (names []string, err error) {
	if len(spriteInfo) == 0 {
		if err = alfred.LoadJSON(path.Join(workflow.WorkflowDir(), "emoji.json"), &spriteInfo); err != nil {
			return
		}
	}

	for i := range spriteInfo {
		names = append(names, spriteInfo[i].Name)
	}

	return
}

func fileExists(filename string) (exists bool) {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}
