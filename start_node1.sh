#!/bin/bash

NODE1_HOME="$HOME/.evmd_node1"

echo "🚀 Starting Node 1..."
echo "RPC: http://localhost:26657"
echo "JSON-RPC: http://localhost:8545"
echo "Press Ctrl+C to stop"
echo ""

evmd start \
    --home "$NODE1_HOME" \
    --log_level info \
    --minimum-gas-prices=0.0001atest \
    --json-rpc.api eth,txpool,personal,net,debug,web3 \
    --chain-id 9001
