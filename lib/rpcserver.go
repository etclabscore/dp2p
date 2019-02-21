package lib

import (
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

func Run() {

	lg := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	lg.Verbosity(log.Lvl(9))
	log.Root().SetHandler(lg)

	apis := []rpc.API{
		rpc.API{
			Namespace: "disco",
			Version:   "2.0",
			Service:   new(DiscoP2PService),
			Public:    true,
		},
	}
	endpoint := ":8544" // TODO flag me
	modules := []string{"disco"}
	cors := []string{"*"}
	vhosts := []string{"*"}

	_, _, err := rpc.StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts, rpc.DefaultHTTPTimeouts)
	if err != nil {
		log.Crit("http endpoint", "error", err)
	}
	log.Info("http endpoint opened", "endpoint", endpoint, "cors", cors, "vhosts", vhosts)
	select {}
}
