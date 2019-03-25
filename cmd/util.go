package cmd

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"log"
	"os"
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