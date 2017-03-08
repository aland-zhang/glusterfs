package gfid

import (
	"os"
	"time"
)

type InodeInfo struct {
	GFID   string      `json:"gfid,omitempty"`
	Path   string      `json:"path,omitempty"`
	Name   string      `json:"name,omitempty"`   // base name of the file
	Size   int64       `json:"size,omitempty"`   // length in bytes for regular files; system-dependent for others
	Mode   os.FileMode `json:"mode,omitempty"`   // file mode bits
	Mtime  time.Time   `json:"mtime,omitempty"`  // modification time
	IsDir  bool        `json:"is_dir,omitempty"` // abbreviation for Mode().IsDir()
	MD5Sum string      `json:"md5sum,omitempty"` // base name of the file
}
