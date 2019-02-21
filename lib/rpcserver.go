package lib

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/ethereum/go-ethereum/log"
)

func Run() {
	// service := new(somelib.Struct)
	// err := rpc.Register(service)
	// // handle error
	ds := new(DiscoP2PService)
	rpc.Register(ds)
	rpc.HandleHTTP()

	lg := log.New("rpcsrv")
	listening, err := net.Listen("tcp", ":1234")
	if err != nil {
		lg.Crit("could not start listening", "error", err)
	}
	lg.Info("serving", "port", ":1234")
	err = http.Serve(listening, nil)
	if err != nil {
		lg.Crit("could not serve http", "error", err)
	}
}
