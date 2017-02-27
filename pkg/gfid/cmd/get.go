package cmd

import (
	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdGet() *cobra.Command {
	storageDir, err := GetStorageDir()
	if err != nil {
		log.Fatalln(err)
	}

	var gfids []string
	cmd := &cobra.Command{
		Use: "get",
		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, "gfids")
			err := lib.Get(storageDir, gfids)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&storageDir, "storage-dir", storageDir, "Directory where LevelDB file is stored.")
	cmd.Flags().StringSliceVar(&gfids, "gfids", []string{}, "Comma separated list of GFIDs.")
	return cmd
}
