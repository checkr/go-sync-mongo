package cmd

import (
	"fmt"
	s "strings"

	db "../db"
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
			Database:    viper.GetString("src-db"),
			Collections: s.Split(viper.GetString("src-collections"), ","),
		}
		src, err := db.NewConnection(srcConfig)
		if err != nil {
			fmt.Errorf("Error: %s", err)
		}

		dstConfig := db.Config{
			URI: viper.GetString("dst"),
			SSL: viper.GetBool("dst-ssl"),
			Creds: mgo.Credential{
				Username: viper.GetString("dst-username"),
				Password: viper.GetString("dst-password"),
			},
			Database:    viper.GetString("dst-db"),
			Collections: make([]string, 0),
		}
		dst, err := db.NewConnection(dstConfig)
		if err != nil {
			fmt.Errorf("Error: %s", err)
		}

		err = src.SyncOplog(dst)
		if err != nil {
			fmt.Println(err)
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
