package lib

import (
	"encoding/hex"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/syndtr/goleveldb/leveldb"
	"k8s.io/kubernetes/pkg/util/json"
)

func Get(storageDir string, gfids []string) error {
	db, err := leveldb.OpenFile(storageDir, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"GFID", "SIZE", "MODE", "MTIME", "PATH"})
	for _, gfid := range gfids {
		gfid = strings.Replace(gfid, "-", "", -1)

		key, err := hex.DecodeString(gfid)
		if err != nil {
			return err
		}

		bytes, err := db.Get(key, nil)
		if err != nil {
			table.Append([]string{
				gfid,
				"",
				"",
				"",
				"<NOT FOUND>",
			})
		} else {
			var ii InodeInfo
			err = json.Unmarshal(bytes, &ii)
			if err != nil {
				return err
			}

			table.Append([]string{
				ii.GFID,
				humanize.Bytes(uint64(ii.Size)),
				ii.Mode.String(),
				ii.Mtime.String(),
				ii.Path,
			})
		}
	}
	table.Render()
	return nil
}
