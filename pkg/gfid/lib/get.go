package lib

import (
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

		bytes, err := db.Get([]byte(gfid), nil)
		if err != nil {
			return err
		}

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
	table.Render()
	return nil
}
