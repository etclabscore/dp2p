package lib

import (
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/ethereum/go-ethereum/log"
)

func Run() {
	// service := new(somelib.Struct)
	// err := rpc.Register(service)
	// // handle error
	ds := new(DiscoP2PService)
	rpc.Register(ds)
	rpc.HandleHTTP()

	addr := ":8544"

	lg := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	lg.Verbosity(log.Lvl(9))
	log.Root().SetHandler(lg)

	listening, err := net.Listen("tcp", addr)
	if err != nil {
		log.Crit("could not start listening", "error", err)
	}
	log.Info("serving", "addr", addr)
	err = http.Serve(listening, nil)
	if err != nil {
		log.Crit("could not serve http", "error", err)
	}
}
