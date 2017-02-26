package cmd

import (
	"fmt"
	"os"

	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdGet() *cobra.Command {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	gfidFile := os.TempDir() + "/" + hostname + ".gfid"
	fmt.Println("____________", gfidFile)

	var gfids []string
	cmd := &cobra.Command{
		Use: "get",
		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, "gfids")
			err := lib.Get(gfidFile, gfids)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&gfidFile, "gfid-file", gfidFile, "Path to .gfid file")
	cmd.Flags().StringSliceVar(&gfids, "gfids", []string{}, "Comma separated list of GFIDs.")
	return cmd
}
