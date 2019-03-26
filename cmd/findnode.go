// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"github.com/etclabscore/dp2p/discover"
	"github.com/spf13/cobra"
	"log"
	"net"
	"time"
)

// findnodeCmd represents the neighbors command
var findnodeCmd = &cobra.Command{
	Use:   "findnode",
	Short: "Send a devp2p FINDNODE request to an enode (with preliminary PING/PONG)",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {

		en := mustEnodeArg(args)

		discover.SetResponseTimeout(time.Duration(int32(respTimeout)) * time.Millisecond)

		u := mustUdp()

		nodes, err := u.Findnode(en.ID(), &net.UDPAddr{IP: en.IP(), Port: en.UDP()}, en.Pubkey())
		if err != nil {
			log.Fatalln(err)
		}

		for _, n := range nodes {
			fmt.Println(n.String())
		}
	},
}

func init() {
	findnodeCmd.PersistentFlags().StringVarP(&listenAddr, "listenaddr", "a", ":30301", "address:port to listen at")
	findnodeCmd.PersistentFlags().IntVarP(&respTimeout, "resptimeout", "t", 500, "milliseconds for devp2p response timeout allowance")
	rootCmd.AddCommand(findnodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// findnodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// findnodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
