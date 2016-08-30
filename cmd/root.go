// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "go-sync-mongo",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-sync-mongo.yaml)")
	RootCmd.PersistentFlags().String("src", "", "mongodb://host1:27017")
	RootCmd.PersistentFlags().String("src-username", "", "source database username")
	RootCmd.PersistentFlags().String("src-password", "", "source database password")
	RootCmd.PersistentFlags().Bool("src-ssl", true, "source ssl enabled (true)")
	RootCmd.PersistentFlags().String("dst", "", "mongodb://host1:27017,host2:27017")
	RootCmd.PersistentFlags().String("dst-username", "", "destination database username")
	RootCmd.PersistentFlags().String("dst-password", "", "destiantion database password")
	RootCmd.PersistentFlags().Bool("dst-ssl", true, "destination ssl enabled (true)")

	viper.BindPFlag("src", RootCmd.PersistentFlags().Lookup("src"))
	viper.BindPFlag("src-username", RootCmd.PersistentFlags().Lookup("src-username"))
	viper.BindPFlag("src-password", RootCmd.PersistentFlags().Lookup("src-password"))
	viper.BindPFlag("src-ssl", RootCmd.PersistentFlags().Lookup("src-ssl"))
	viper.BindPFlag("dst", RootCmd.PersistentFlags().Lookup("dst"))
	viper.BindPFlag("dst-username", RootCmd.PersistentFlags().Lookup("dst-username"))
	viper.BindPFlag("dst-password", RootCmd.PersistentFlags().Lookup("dst-password"))
	viper.BindPFlag("dst-ssl", RootCmd.PersistentFlags().Lookup("dst-ssl"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}
	viper.SetConfigName(".go-sync-mongo") // name of config file (without extension)
	viper.AddConfigPath("$HOME")          // adding home directory as first search path
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("GSM")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
