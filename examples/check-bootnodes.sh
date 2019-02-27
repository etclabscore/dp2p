#!/usr/bin/env bash

# Checks bootnode(enode) reachability for networks supported by a geth client, eg. multi-geth.
# Enodes are each pinged in their own processes, so this script should only take approximately as long as the
# timeout limit to run.
#
# Use:
#  cd $GOPATH/src/github.com/etclabscore/dp2p &&
#  ./check-bootnodes.sh [|network names...]
#
# If no network names are passed, then all available are used.
# - See 'data_dir' var below if you want to change where resulting data goes.
# - See 'timeout_lim' var if you want to change how long (in seconds) before ping attempt resigns.

timeout_lim=$((60*60)) # m * n_seconds
data_dir="data/$(date +%s)-${timeout_lim}s"
all_networks=(mainnet testnet classic social ethersocial mix rinkeby kotti goerli)


set -e


# Build executables geth and dp2p, we'll remove them after use.
curd="$(pwd)"
pushd $GOPATH/src/github.com/ethereum/go-ethereum && make && cp ./build/bin/geth "$curd/geth" && popd
unset curd
go build -o dp2p

trap "rm geth dp2p" EXIT # Remove executable bins on exit.

mkdir -p "$data_dir"

networks=()
for a in "$@"; do
    networks+=("$a")
done
# Set default network list in case none are argued.
if [[ "${#networks[@]}" -eq 0 ]]; then networks=("${all_networks[@]}"); fi


# Collect lists of bootnodes from geth's 'dumpconfig' command.
for net in "${networks[@]}"
do
    [[ -z "$net" ]] && echo 'empty net, continuing' && continue

    echo "Found chain: $net"
    mkdir -p "$data_dir/$net"

    nf="--$net"
    [[ "$net" == mainnet ]] && nf=

    ./geth version > "$data_dir/$net/geth.version" 2>&1

    ./geth 2>/dev/null $nf dumpconfig |
        grep -E "^BootstrapNodes " |
        cut -d'=' -f2- |
        sed 's/\[//g;s/\]//g;s/,/\n/g;s/"//g' |
        sed 's/^ *//g' |
        tee "$data_dir/$net/bootnodes.list"

    while read -r enode; do
        enode_id="$(echo $enode | cut -d'/' -f2- | cut -d '@' -f1)"
        (
            set +e # Must allow dp2p to 'fail', ie exit w/ 1
            echo "Running $net $enode"
            start="$(date +%s)"
            # Python here asks the OS for an open port on the machine.
            ./dp2p addpeer -a ":$(python -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')" -t $timeout_lim "$enode" > "$data_dir/$net$enode_id.log" 2>&1
            res="$?"
            end="$(date +%s)"
            echo "$res $net-$enode "'('$((end-start))'s)' >> "$data_dir/$net/outcomes"
        )&
    done < "$data_dir/$net/bootnodes.list"
done

js="$(jobs -p)"
echo "Backgrounded jobs numberof=$(wc -l <<< $js)"
echo "Backgrounded jobs pids=[ $(printf '%s ' $js)]"

wait

echo Done
echo
echo Outcomes:
for d in $data_dir/*; do
    oks=$(cat "$d/outcomes" | grep '^0' | wc -l)
    fails=$(cat "$d/outcomes" | grep '^1' | wc -l)
    echo -e "$d\tok=$oks\tfails=$fails"
done
