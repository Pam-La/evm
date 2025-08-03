#!/bin/bash

CHAINID="${CHAIN_ID:-9001}"
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
BASEFEE=10000000

# 노드별 디렉토리 설정
NODE1_HOME="$HOME/.evmd_node1"
NODE2_HOME="$HOME/.evmd_node2"

# 포트 설정
NODE1_RPC_PORT=26657
NODE1_P2P_PORT=26656
NODE1_JSON_RPC_PORT=8545

NODE2_RPC_PORT=26667
NODE2_P2P_PORT=26666
NODE2_JSON_RPC_PORT=8555

# Parse input flags
overwrite=""
while [[ $# -gt 0 ]]; do
	key="$1"
	case $key in
	-y)
		overwrite="y"
		shift
		;;
	*)
		echo "Unknown flag: $key"
		exit 1
		;;
	esac
done

# 기존 디렉토리 제거 확인
if [[ $overwrite == "" ]]; then
	if [ -d "$NODE1_HOME" ] || [ -d "$NODE2_HOME" ]; then
		echo "Existing node directories found. Overwrite? [y/n]"
		read -r overwrite
	else
		overwrite="y"
	fi
fi

if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	echo "🧹 Cleaning existing directories..."
	rm -rf "$NODE1_HOME" "$NODE2_HOME"

	echo "🔧 Building evmd..."
	cd evmd && make install && cd ..

	# Validator 키 생성
	VAL1_KEY="validator1"
	VAL1_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"
	
	VAL2_KEY="validator2"
	VAL2_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

	# User 키들
	USER1_KEY="dev0"
	USER1_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

	echo "🏠 Setting up Node 1..."
	# Node 1 설정
	evmd config set client chain-id "$CHAINID" --home "$NODE1_HOME"
	evmd config set client keyring-backend "$KEYRING" --home "$NODE1_HOME"
	
	echo "$VAL1_MNEMONIC" | evmd keys add "$VAL1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE1_HOME"
	echo "$VAL2_MNEMONIC" | evmd keys add "$VAL2_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE1_HOME"
	echo "$USER1_MNEMONIC" | evmd keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE1_HOME"

	evmd init node1 --chain-id "$CHAINID" --home "$NODE1_HOME"

	echo "🏠 Setting up Node 2..."
	# Node 2 설정
	evmd config set client chain-id "$CHAINID" --home "$NODE2_HOME"
	evmd config set client keyring-backend "$KEYRING" --home "$NODE2_HOME"
	
	echo "$VAL1_MNEMONIC" | evmd keys add "$VAL1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE2_HOME"
	echo "$VAL2_MNEMONIC" | evmd keys add "$VAL2_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE2_HOME"
	echo "$USER1_MNEMONIC" | evmd keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$NODE2_HOME"

	evmd init node2 --chain-id "$CHAINID" --home "$NODE2_HOME"

	echo "⚙️ Configuring genesis..."
	# Node 1에서 genesis 설정
	GENESIS1="$NODE1_HOME/config/genesis.json"
	TMP_GENESIS1="$NODE1_HOME/config/tmp_genesis.json"

	# Genesis 파라미터 설정
	jq '.app_state["staking"]["params"]["bond_denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["gov"]["params"]["expedited_min_deposit"][0]["denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["evm"]["params"]["evm_denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["mint"]["params"]["mint_denom"]="atest"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.app_state["consensus"]["params"]["block"]["max_gas"]="1000000000"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"
	jq '.consensus.params.block.max_gas="10000000000"' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"

	# EVM precompiles 활성화
	jq '.app_state["evm"]["params"]["active_static_precompiles"]=["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805"]' "$GENESIS1" >"$TMP_GENESIS1" && mv "$TMP_GENESIS1" "$GENESIS1"

	# Genesis accounts 추가
	evmd genesis add-genesis-account "$VAL1_KEY" 100000000000000000000000000atest --keyring-backend "$KEYRING" --home "$NODE1_HOME"
	evmd genesis add-genesis-account "$VAL2_KEY" 100000000000000000000000000atest --keyring-backend "$KEYRING" --home "$NODE1_HOME"
	evmd genesis add-genesis-account "$USER1_KEY" 1000000000000000000000atest --keyring-backend "$KEYRING" --home "$NODE1_HOME"

	echo "🔐 Creating validator transactions..."
	# Validator 1 gentx 생성
	evmd genesis gentx "$VAL1_KEY" 1000000000000000000000atest --gas-prices ${BASEFEE}atest --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$NODE1_HOME"

	# Validator 2 gentx 생성 (Node 2에서)
	cp "$GENESIS1" "$NODE2_HOME/config/genesis.json"
	evmd genesis gentx "$VAL2_KEY" 1000000000000000000000atest --gas-prices ${BASEFEE}atest --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$NODE2_HOME"

	# Node 2의 gentx를 Node 1으로 복사
	cp "$NODE2_HOME/config/gentx/"* "$NODE1_HOME/config/gentx/"

	# Genesis transactions 수집
	evmd genesis collect-gentxs --home "$NODE1_HOME"

	# 최종 genesis를 Node 2에 복사
	cp "$NODE1_HOME/config/genesis.json" "$NODE2_HOME/config/genesis.json"

	echo "🌐 Configuring network settings..."
	# Config 파일 설정
	CONFIG1="$NODE1_HOME/config/config.toml"
	CONFIG2="$NODE2_HOME/config/config.toml"
	APP_TOML1="$NODE1_HOME/config/app.toml"
	APP_TOML2="$NODE2_HOME/config/app.toml"

	# Node 1 포트 설정
	sed -i.bak "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:$NODE1_RPC_PORT\"/g" "$CONFIG1"
	sed -i.bak "s/laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$NODE1_P2P_PORT\"/g" "$CONFIG1"
	sed -i.bak "s/address = \"127.0.0.1:8545\"/address = \"127.0.0.1:$NODE1_JSON_RPC_PORT\"/g" "$APP_TOML1"

	# Node 2 포트 설정
	sed -i.bak "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:$NODE2_RPC_PORT\"/g" "$CONFIG2"
	sed -i.bak "s/laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$NODE2_P2P_PORT\"/g" "$CONFIG2"
	sed -i.bak "s/address = \"127.0.0.1:8545\"/address = \"127.0.0.1:$NODE2_JSON_RPC_PORT\"/g" "$APP_TOML2"

	# Node 1의 peer ID 가져오기
	NODE1_ID=$(evmd cometbft show-node-id --home "$NODE1_HOME")
	NODE2_ID=$(evmd cometbft show-node-id --home "$NODE2_HOME")

	# P2P 연결 설정
	sed -i.bak "s/persistent_peers = \"\"/persistent_peers = \"$NODE2_ID@127.0.0.1:$NODE2_P2P_PORT\"/g" "$CONFIG1"
	sed -i.bak "s/persistent_peers = \"\"/persistent_peers = \"$NODE1_ID@127.0.0.1:$NODE1_P2P_PORT\"/g" "$CONFIG2"

	# 블록 시간 설정
	sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "500ms"/g' "$CONFIG1"
	sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "500ms"/g' "$CONFIG2"

	# mempool 설정
	sed -i.bak 's/size = 5000/size = 10000/g' "$CONFIG1"
	sed -i.bak 's/size = 5000/size = 10000/g' "$CONFIG2"

	# API 활성화
	sed -i.bak 's/enable = false/enable = true/g' "$APP_TOML1"
	sed -i.bak 's/enable = false/enable = true/g' "$APP_TOML2"

	echo "✅ Multi-node setup completed!"
	echo ""
	echo "🚀 To start the nodes:"
	echo "Terminal 1: evmd start --home $NODE1_HOME --log_level info --minimum-gas-prices=0.0001atest --json-rpc.api eth,txpool,personal,net,debug,web3"
	echo "Terminal 2: evmd start --home $NODE2_HOME --log_level info --minimum-gas-prices=0.0001atest --json-rpc.api eth,txpool,personal,net,debug,web3"
	echo ""
	echo "🔗 RPC Endpoints:"
	echo "Node 1: http://localhost:$NODE1_RPC_PORT (JSON-RPC: http://localhost:$NODE1_JSON_RPC_PORT)"
	echo "Node 2: http://localhost:$NODE2_RPC_PORT (JSON-RPC: http://localhost:$NODE2_JSON_RPC_PORT)"
	echo ""
	echo "🔍 Node IDs:"
	echo "Node 1: $NODE1_ID"
	echo "Node 2: $NODE2_ID"
fi
