package cmd

import (
	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdGet() *cobra.Command {
	storageDir, err := GetStorageDir()
	if err != nil {
		log.Fatalln(err)
	}

	var vol string
	var gfids []string
	cmd := &cobra.Command{
		Use: "get",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if vol != "" {
				gfids, err = lib.GetGFIDs(vol)
				if err != nil {
					log.Fatalln(err)
				}
			}
			err = lib.Get(storageDir, gfids)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&storageDir, "storage-dir", storageDir, "Directory where LevelDB file is stored.")
	cmd.Flags().StringSliceVar(&gfids, "gfids", []string{}, "Comma separated list of GFIDs.")
	cmd.Flags().StringVarP(&vol, "volume", "v", "", "Volume name used to detect split-brain gfids")
	return cmd
}
