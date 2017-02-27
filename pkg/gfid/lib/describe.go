package lib

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

//  gluster volume info vol
// trusted.afr.vol-client-0
// trusted.afr.<volname>-client-<subvolume-index>

var brickID = regexp.MustCompile(`^Brick[0-9]+:.*$`)

func ListSubvolumes(vol string) ([]string, error) {
	cmdOut, err := exec.Command("gluster", "volume", "info", vol).Output()
	if err != nil {
		return nil, err
	}

	sub := make([]string, 0)
	lines := strings.Split(string(cmdOut), "\n")
	for _, line := range lines {
		if brickID.MatchString(line) {
			sub = append(sub, line[strings.Index(line, " ")+1:])
		}
	}
	return sub, nil
}

func PrintSubvolumes(vol string) error {
	subs, err := ListSubvolumes(vol)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"INDEX", "BRICK"})
	for i, sub := range subs {
		table.Append([]string{
			strconv.Itoa(i),
			sub,
		})
	}
	table.Render()
	return nil
}
