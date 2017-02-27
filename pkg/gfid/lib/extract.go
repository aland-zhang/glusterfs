package lib

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/appscode/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"golang.org/x/sys/unix"
	"k8s.io/kubernetes/pkg/util/json"
)

func Extract(dataDir, storageDir string, computeChecksum bool) error {
	db, err := leveldb.OpenFile(storageDir, nil)
	if err != nil {
		return err
	}

	filepath.Walk(dataDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := fi.Name()

		if path == dataDir+"/.glusterfs" ||
			path == dataDir+"/.trashcan" ||
			strings.HasPrefix(path, dataDir+"/.glusterfs/") ||
			strings.HasPrefix(path, dataDir+"/.trashcan") {
			log.Infoln("Skipping path", path)
			return nil
		}
		gfid := make([]byte, 16)
		sz, err := unix.Getxattr(path, "trusted.gfid", gfid)
		if err != nil {
			return err
		}
		gfid = gfid[:sz]

		ii := &InodeInfo{
			GFID:  hex.EncodeToString(gfid),
			Path:  path,
			Name:  name,
			Size:  fi.Size(),
			Mode:  fi.Mode(),
			Mtime: fi.ModTime(),
			IsDir: fi.IsDir(),
		}
		if !ii.IsDir && computeChecksum {
			hasher := md5.New()
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(hasher, f); err != nil {
				return err
			}
			ii.MD5Sum = hex.EncodeToString(hasher.Sum(nil))
		}
		fmt.Println("gfid:", ii.GFID)

		bytes, err := json.Marshal(ii)
		if err != nil {
			return err
		}
		return db.Put(gfid, bytes, &opt.WriteOptions{Sync: true})
	})

	return db.Close()
}
