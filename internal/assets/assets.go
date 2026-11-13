// Package assets embeds the app's UI assets.
package assets

import (
	"embed"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

// icons holds the protocol PNGs reused from the old project plus a folder
// SVG. Lookup is by file name: mysql.png serves protocol "mysql".
//
//go:embed icons/*
var icons embed.FS

var (
	iconsOnce sync.Once
	bitmaps   map[string]*ui.Bitmap
	svgs      map[string]*ui.SVG
)

func loadIcons() {
	bitmaps = map[string]*ui.Bitmap{}
	svgs = map[string]*ui.SVG{}
	entries, err := icons.ReadDir("icons")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		ext := filepath.Ext(name)
		data, err := icons.ReadFile("icons/" + entry.Name())
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(name, ext)
		switch ext {
		case ".png":
			if b, err := ui.DecodeBitmap(data); err == nil {
				bitmaps[stem] = b
			}
		case ".svg":
			if s, err := ui.ParseSVG(data); err == nil {
				svgs[stem] = s
			}
		}
	}
}

// ProtocolBitmap returns the icon of a connection protocol. SSH shows the
// Linux icon and WinRM the Windows icon, matching the old app's mapping.
func ProtocolBitmap(protocol string) *ui.Bitmap {
	iconsOnce.Do(loadIcons)
	key := strings.ToLower(strings.TrimSpace(protocol))
	switch key {
	case "", "ssh":
		key = "linux"
	case "winrm":
		key = "windows"
	}
	return bitmaps[key]
}

// FolderSVG returns the folder icon.
func FolderSVG() *ui.SVG {
	iconsOnce.Do(loadIcons)
	return svgs["folder"]
}

// ObjectSVG returns one of the database object icons: table, view or field.
func ObjectSVG(name string) *ui.SVG {
	iconsOnce.Do(loadIcons)
	return svgs[name]
}
