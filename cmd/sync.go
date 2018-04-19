package cmd

import (
	"fmt"
	"os"

	db "github.com/checkr/go-sync-mongo/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Tails the source oplog and syncs to destination",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		srcConfig := db.Config{
			URI: viper.GetString("src"),
			SSL: viper.GetBool("src-ssl"),
			Creds: mgo.Credential{
				Username: viper.GetString("src-username"),
				Password: viper.GetString("src-password"),
			},
		}
		src, err := db.NewConnection(srcConfig)
		if err != nil {
			fmt.Printf("Error: new src connection - %s\n", err)
			os.Exit(1)
		}

		dstConfig := db.Config{
			URI: viper.GetString("dst"),
			SSL: viper.GetBool("dst-ssl"),
			Creds: mgo.Credential{
				Username: viper.GetString("dst-username"),
				Password: viper.GetString("dst-password"),
			},
		}
		dst, err := db.NewConnection(dstConfig)
		if err != nil {
			fmt.Printf("Error: new dst connection - %s\n", err)
			os.Exit(1)
		}

		err = src.SyncOplog(dst)
		if err != nil {
			fmt.Printf("Error: sync oplog - %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Int32("since", 0, "seconds since the Unix epoch")
	syncCmd.Flags().Int32("ordinal", 0, "incrementing ordinal for operations within a given second")
	viper.BindPFlag("since", syncCmd.Flags().Lookup("since"))
	viper.BindPFlag("ordinal", syncCmd.Flags().Lookup("ordinal"))
}
