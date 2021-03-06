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
	"log"
	"net"
	"time"

	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send a PING request to a given enode",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {

		en := mustEnodeArg(args)

		discover.SetResponseTimeout(time.Duration(int32(respTimeout)) * time.Millisecond)

		u := mustUdp()

		tstart := time.Now()

		err := <-u.SendPing(en.ID(), &net.UDPAddr{IP: en.IP(), Port: en.UDP()}, func() {
			fmt.Println("pong", time.Since(tstart))
		})
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	pingCmd.PersistentFlags().StringVarP(&listenAddr, "listenaddr", "a", ":30301", "address:port to listen at")
	pingCmd.PersistentFlags().IntVarP(&respTimeout, "resptimeout", "t", 500, "milliseconds for devp2p response timeout allowance")
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
