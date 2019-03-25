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
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/spf13/cobra"
	"log"
	"net"
)

// neighborsCmd represents the neighbors command
var neighborsCmd = &cobra.Command{
	Use:   "neighbors",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		en := mustEnodeArg(args)

		nodeKey, _ := crypto.GenerateKey()

		addr, err := net.ResolveUDPAddr("udp", listenAddr)
		if err != nil {
			utils.Fatalf("-ResolveUDPAddr: %v", err)
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			utils.Fatalf("-ListenUDP: %v", err)
		}

		db, _ := enode.OpenDB("")
		ln := enode.NewLocalNode(db, nodeKey)
		cfg := discover.Config{
			PrivateKey:  nodeKey,
			//NetRestrict: restrictList,
		}
		_, u, err := discover.ListenUDP(conn, ln, cfg)
		if err != nil {
			utils.Fatalf("%v", err)
		}
		nodes, err := u.Findnode(en.ID(), &net.UDPAddr{IP: en.IP(), Port: en.UDP()}, en.Pubkey())

		if err != nil {
			log.Fatalln(err)
		}
		log.Println("got neighbors ok:", len(nodes))
		for _, n := range nodes {
			fmt.Println(n.String())
		}
	},
}

func init() {
	neighborsCmd.PersistentFlags().StringVarP(&listenAddr, "listenaddr", "a", ":30301", "address:port to listen at")

	rootCmd.AddCommand(neighborsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// neighborsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// neighborsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
