package lib

import (
	"encoding/hex"
	"os"
	"os/exec"
	"strings"

	"github.com/appscode/go/net"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/syndtr/goleveldb/leveldb"
	"k8s.io/kubernetes/pkg/util/json"
)

func GetGFIDs(vol string) ([]string, error) {
	_, ip, err := net.NodeIP()
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	cmdOut, err := exec.Command("gluster", "volume", "heal", vol, "info").Output() // , "split-brain"
	if err != nil {
		return nil, err
	}

	gfids := make([]string, 0)
	found := false

	lines := strings.Split(string(cmdOut), "\n")
	for _, line := range lines {
		// fmt.Println(line)
		if strings.HasPrefix(line, "Brick ") {
			l := line[len("Brick "):]
			d := strings.Split(l, ":")
			found = (d[0] == ip.String() || d[0] == hostname)
			continue
		} else if found && strings.HasPrefix(line, "<gfid:") {
			l := line[len("<gfid:"):strings.Index(line, ">")]
			gfids = append(gfids, l)
		}
	}
	return gfids, nil
}

func Get(storageDir string, gfids []string) error {
	db, err := leveldb.OpenFile(storageDir, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"GFID", "MODE", "SIZE", "MTIME", "PATH"})
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
				ii.Mode.String(),
				humanize.Bytes(uint64(ii.Size)),
				ii.Mtime.String(),
				ii.Path,
			})
		}
	}
	table.Render()
	return nil
}
