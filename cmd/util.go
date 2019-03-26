package cmd

import (
	"github.com/etclabscore/dp2p/discover"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"log"
	"net"
	"os"
)

var (
	connectTimeout int
	respTimeout int
	listenAddr string
)

func mustEnodeArg(args []string) *enode.Node {
	if len(args) == 0 {
		log.Println("need enode as first argument")
		os.Exit(1)
	}
	eni := args[0]
	en, err := enode.ParseV4(eni)
	if err != nil {
		log.Println("malformed enode", eni)
		os.Exit(1)
	}
	return en
}

func mustUdp() *discover.Udp {
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
	return u
}