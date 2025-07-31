#!/bin/bash

NODE2_HOME="$HOME/.evmd_node2"

echo "🚀 Starting Node 2..."
echo "RPC: http://localhost:26667"
echo "JSON-RPC: http://localhost:8555"
echo "Press Ctrl+C to stop"
echo ""

evmd start \
    --home "$NODE2_HOME" \
    --log_level info \
    --minimum-gas-prices=0.0001atest \
    --json-rpc.api eth,txpool,personal,net,debug,web3 \
    --chain-id 9001
