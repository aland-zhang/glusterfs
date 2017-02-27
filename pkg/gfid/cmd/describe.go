package cmd

import (
	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdDescribe() *cobra.Command {
	var vol string
	cmd := &cobra.Command{
		Use: "describe",
		Run: func(cmd *cobra.Command, args []string) {
			err := lib.PrintSubvolumes(vol)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&vol, "volume", "", "Volume name used to detect split-brain gfids eg, vol")
	return cmd
}
