#!/bin/bash

CHAINID="swisstronik_1291-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
HOMEDIR="$HOME/.swisstronik"
BINARY="./build/swisstronikd"

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# Compliance proxy contract
COMPLIANCE_PROXY_BYTECODE="608060405234801561001057600080fd5b50600436106100935760003560e01c806395183eb71161006657806395183eb714610134578063ace417e014610164578063d832c2f014610194578063d916d4e2146101c4578063e711d86d146101e257610093565b80633fccb7481461009857806344e48ac6146100b65780636be006d4146100d45780638cc2551d14610104575b600080fd5b6100a06101fe565b6040516100ad9190610c60565b60405180910390f35b6100be6102c7565b6040516100cb9190610c60565b60405180910390f35b6100ee60048036038101906100e99190610cf4565b610390565b6040516100fb9190610c60565b60405180910390f35b61011e60048036038101906101199190610d57565b6105e2565b60405161012b9190610c60565b60405180910390f35b61014e60048036038101906101499190610cf4565b6107bd565b60405161015b919061101c565b60405180910390f35b61017e60048036038101906101799190610cf4565b6108b5565b60405161018b9190611059565b60405180910390f35b6101ae60048036038101906101a991906111bc565b6109bd565b6040516101bb9190611059565b60405180910390f35b6101cc610ac1565b6040516101d99190611227565b60405180910390f35b6101fc60048036038101906101f791906112f7565b610ac6565b005b60606000604051602401604051602081830303815290604052633db94a0460e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610279919061137c565b600060405180830381855afa9150503d80600081146102b4576040519150601f19603f3d011682016040523d82523d6000602084013e6102b9565b606091505b509150915080935050505090565b6060600060405160240160405160208183030381529060405263d0376bd260e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610342919061137c565b600060405180830381855afa9150503d806000811461037d576040519150601f19603f3d011682016040523d82523d6000602084013e610382565b606091505b509150915080935050505090565b60606000600167ffffffffffffffff8111156103af576103ae611079565b5b6040519080825280601f01601f1916602001820160405280156103e15781602001600182028036833780820191505090505b50905060006040518060400160405280600c81526020017f636861696e5f313239312d310000000000000000000000000000000000000000815250905060006040518060400160405280600681526020017f736368656d610000000000000000000000000000000000000000000000000000815250905060006040518060400160405280601481526020017f697373756572566572696669636174696f6e4964000000000000000000000000815250905060008087856002640100000000426104aa91906113cc565b60008a8989896040516024016104c89998979695949392919061149b565b60405160208183030381529060405263e62364ab60e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610539919061137c565b6000604051808303816000865af19150503d8060008114610576576040519150601f19603f3d011682016040523d82523d6000602084013e61057b565b606091505b509150915060008180602001905181019061059691906115b4565b90507f109923c5f814b7944f9591557d6a8d1dde907b5cc4cebccf84d59959347589bc83826040516105c99291906115fd565b60405180910390a1809950505050505050505050919050565b60606000600167ffffffffffffffff81111561060157610600611079565b5b6040519080825280601f01601f1916602001820160405280156106335781602001600182028036833780820191505090505b50905060006040518060400160405280601481526020017f697373756572566572696669636174696f6e496400000000000000000000000081525090506000808660026401000000004261068791906113cc565b60008787878c6040516024016106a49897969594939291906116d4565b60405160208183030381529060405263c220658060e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610715919061137c565b6000604051808303816000865af19150503d8060008114610752576040519150601f19603f3d011682016040523d82523d6000602084013e610757565b606091505b509150915060008180602001905181019061077291906115b4565b90507f109923c5f814b7944f9591557d6a8d1dde907b5cc4cebccf84d59959347589bc83826040516107a59291906115fd565b60405180910390a18097505050505050505092915050565b6060600082306040516024016107d4929190611788565b60405160208183030381529060405263cc8995ec60e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610845919061137c565b600060405180830381855afa9150503d8060008114610880576040519150601f19603f3d011682016040523d82523d6000602084013e610885565b606091505b5091509150606082156108a957818060200190518101906108a69190611b00565b90505b80945050505050919050565b6000606060008360026000846040516024016108d49493929190611bf8565b604051602081830303815290604052634887fcd860e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610945919061137c565b600060405180830381855afa9150503d8060008114610980576040519150601f19603f3d011682016040523d82523d6000602084013e610985565b606091505b509150915081156109af57808060200190518101906109a49190611c70565b9450505050506109b8565b60009450505050505b919050565b6000808360026000856040516024016109d99493929190611bf8565b604051602081830303815290604052634887fcd860e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610a4a919061137c565b600060405180830381855afa9150503d8060008114610a85576040519150601f19603f3d011682016040523d82523d6000602084013e610a8a565b606091505b50915091508115610ab35780806020019051810190610aa99190611c70565b9350505050610abb565b600093505050505b92915050565b600281565b600081604051602401610ad99190610c60565b60405160208183030381529060405263e711d86d60e01b6020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff8381831617835250505050905060008061040473ffffffffffffffffffffffffffffffffffffffff1683604051610b4a919061137c565b6000604051808303816000865af19150503d8060008114610b87576040519150601f19603f3d011682016040523d82523d6000602084013e610b8c565b606091505b50915091507f1e17dfd999686b132e44592148fe5bc466b7e9997dfbdf8791e92df987182e9f8282604051610bc29291906115fd565b60405180910390a150505050565b600081519050919050565b600082825260208201905092915050565b60005b83811015610c0a578082015181840152602081019050610bef565b60008484015250505050565b6000601f19601f8301169050919050565b6000610c3282610bd0565b610c3c8185610bdb565b9350610c4c818560208601610bec565b610c5581610c16565b840191505092915050565b60006020820190508181036000830152610c7a8184610c27565b905092915050565b6000604051905090565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610cc182610c96565b9050919050565b610cd181610cb6565b8114610cdc57600080fd5b50565b600081359050610cee81610cc8565b92915050565b600060208284031215610d0a57610d09610c8c565b5b6000610d1884828501610cdf565b91505092915050565b6000819050919050565b610d3481610d21565b8114610d3f57600080fd5b50565b600081359050610d5181610d2b565b92915050565b60008060408385031215610d6e57610d6d610c8c565b5b6000610d7c85828601610cdf565b9250506020610d8d85828601610d42565b9150509250929050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600063ffffffff82169050919050565b610ddc81610dc3565b82525050565b600082825260208201905092915050565b6000610dfe82610bd0565b610e088185610de2565b9350610e18818560208601610bec565b610e2181610c16565b840191505092915050565b610e3581610cb6565b82525050565b600081519050919050565b600082825260208201905092915050565b6000610e6282610e3b565b610e6c8185610e46565b9350610e7c818560208601610bec565b610e8581610c16565b840191505092915050565b600061014083016000830151610ea96000860182610dd3565b5060208301518482036020860152610ec18282610df3565b9150506040830151610ed66040860182610e2c565b5060608301518482036060860152610eee8282610e57565b9150506080830151610f036080860182610dd3565b5060a0830151610f1660a0860182610dd3565b5060c083015184820360c0860152610f2e8282610df3565b91505060e083015184820360e0860152610f488282610e57565b915050610100830151848203610100860152610f648282610e57565b915050610120830151610f7b610120860182610dd3565b508091505092915050565b6000610f928383610e90565b905092915050565b6000602082019050919050565b6000610fb282610d97565b610fbc8185610da2565b935083602082028501610fce85610db3565b8060005b8581101561100a5784840389528151610feb8582610f86565b9450610ff683610f9a565b925060208a01995050600181019050610fd2565b50829750879550505050505092915050565b600060208201905081810360008301526110368184610fa7565b905092915050565b60008115159050919050565b6110538161103e565b82525050565b600060208201905061106e600083018461104a565b92915050565b600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6110b182610c16565b810181811067ffffffffffffffff821117156110d0576110cf611079565b5b80604052505050565b60006110e3610c82565b90506110ef82826110a8565b919050565b600067ffffffffffffffff82111561110f5761110e611079565b5b602082029050602081019050919050565b600080fd5b6000611138611133846110f4565b6110d9565b9050808382526020820190506020840283018581111561115b5761115a611120565b5b835b8181101561118457806111708882610cdf565b84526020840193505060208101905061115d565b5050509392505050565b600082601f8301126111a3576111a2611074565b5b81356111b3848260208601611125565b91505092915050565b600080604083850312156111d3576111d2610c8c565b5b60006111e185828601610cdf565b925050602083013567ffffffffffffffff81111561120257611201610c91565b5b61120e8582860161118e565b9150509250929050565b61122181610dc3565b82525050565b600060208201905061123c6000830184611218565b92915050565b600080fd5b600067ffffffffffffffff82111561126257611261611079565b5b61126b82610c16565b9050602081019050919050565b82818337600083830152505050565b600061129a61129584611247565b6110d9565b9050828152602081018484840111156112b6576112b5611242565b5b6112c1848285611278565b509392505050565b600082601f8301126112de576112dd611074565b5b81356112ee848260208601611287565b91505092915050565b60006020828403121561130d5761130c610c8c565b5b600082013567ffffffffffffffff81111561132b5761132a610c91565b5b611337848285016112c9565b91505092915050565b600081905092915050565b600061135682610bd0565b6113608185611340565b9350611370818560208601610bec565b80840191505092915050565b6000611388828461134b565b915081905092915050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b60006113d782611393565b91506113e283611393565b9250826113f2576113f161139d565b5b828206905092915050565b61140681610cb6565b82525050565b600082825260208201905092915050565b600061142882610e3b565b611432818561140c565b9350611442818560208601610bec565b61144b81610c16565b840191505092915050565b6000819050919050565b6000819050919050565b600061148561148061147b84611456565b611460565b610dc3565b9050919050565b6114958161146a565b82525050565b6000610120820190506114b1600083018c6113fd565b81810360208301526114c3818b61141d565b90506114d2604083018a611218565b6114df6060830189611218565b6114ec608083018861148c565b81810360a08301526114fe8187610c27565b905081810360c0830152611512818661141d565b905081810360e0830152611526818561141d565b9050611536610100830184611218565b9a9950505050505050505050565b600061155761155284611247565b6110d9565b90508281526020810184848401111561157357611572611242565b5b61157e848285610bec565b509392505050565b600082601f83011261159b5761159a611074565b5b81516115ab848260208601611544565b91505092915050565b6000602082840312156115ca576115c9610c8c565b5b600082015167ffffffffffffffff8111156115e8576115e7610c91565b5b6115f484828501611586565b91505092915050565b6000604082019050611612600083018561104a565b81810360208301526116248184610c27565b90509392505050565b7f636861696e5f313239312d310000000000000000000000000000000000000000600082015250565b6000611663600c8361140c565b915061166e8261162d565b602082019050919050565b7f736368656d610000000000000000000000000000000000000000000000000000600082015250565b60006116af60068361140c565b91506116ba82611679565b602082019050919050565b6116ce81610d21565b82525050565b6000610140820190506116ea600083018b6113fd565b81810360208301526116fb81611656565b905061170a604083018a611218565b6117176060830189611218565b611724608083018861148c565b81810360a08301526117368187610c27565b905081810360c0830152611749816116a2565b905081810360e083015261175d818661141d565b905061176d610100830185611218565b61177b6101208301846116c5565b9998505050505050505050565b600060408201905061179d60008301856113fd565b6117aa60208301846113fd565b9392505050565b600067ffffffffffffffff8211156117cc576117cb611079565b5b602082029050602081019050919050565b600080fd5b600080fd5b6117f081610dc3565b81146117fb57600080fd5b50565b60008151905061180d816117e7565b92915050565b60008151905061182281610cc8565b92915050565b600067ffffffffffffffff82111561184357611842611079565b5b61184c82610c16565b9050602081019050919050565b600061186c61186784611828565b6110d9565b90508281526020810184848401111561188857611887611242565b5b611893848285610bec565b509392505050565b600082601f8301126118b0576118af611074565b5b81516118c0848260208601611859565b91505092915050565b600061014082840312156118e0576118df6117dd565b5b6118eb6101406110d9565b905060006118fb848285016117fe565b600083015250602082015167ffffffffffffffff81111561191f5761191e6117e2565b5b61192b84828501611586565b602083015250604061193f84828501611813565b604083015250606082015167ffffffffffffffff811115611963576119626117e2565b5b61196f8482850161189b565b6060830152506080611983848285016117fe565b60808301525060a0611997848285016117fe565b60a08301525060c082015167ffffffffffffffff8111156119bb576119ba6117e2565b5b6119c784828501611586565b60c08301525060e082015167ffffffffffffffff8111156119eb576119ea6117e2565b5b6119f78482850161189b565b60e08301525061010082015167ffffffffffffffff811115611a1c57611a1b6117e2565b5b611a288482850161189b565b61010083015250610120611a3e848285016117fe565b6101208301525092915050565b6000611a5e611a59846117b1565b6110d9565b90508083825260208201905060208402830185811115611a8157611a80611120565b5b835b81811015611ac857805167ffffffffffffffff811115611aa657611aa5611074565b5b808601611ab389826118c9565b85526020850194505050602081019050611a83565b5050509392505050565b600082601f830112611ae757611ae6611074565b5b8151611af7848260208601611a4b565b91505092915050565b600060208284031215611b1657611b15610c8c565b5b600082015167ffffffffffffffff811115611b3457611b33610c91565b5b611b4084828501611ad2565b91505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b6000611b818383610e2c565b60208301905092915050565b6000602082019050919050565b6000611ba582611b49565b611baf8185611b54565b9350611bba83611b65565b8060005b83811015611beb578151611bd28882611b75565b9750611bdd83611b8d565b925050600181019050611bbe565b5085935050505092915050565b6000608082019050611c0d60008301876113fd565b611c1a6020830186611218565b611c27604083018561148c565b8181036060830152611c398184611b9a565b905095945050505050565b611c4d8161103e565b8114611c5857600080fd5b50565b600081519050611c6a81611c44565b92915050565b600060208284031215611c8657611c85610c8c565b5b6000611c9484828501611c5b565b9150509291505056fea264697066735822122089e8cb2a9f01a51544fbb4e04f64b388a09ffddddf31b620ca45e818ebe6f56264736f6c63430008180033"
COMPLIANCE_PROXY_CODEHASH="0x92799fbdef25e4617da385123277a186a4b1a1231efbb5f0ad6ae21023c14adc"

# Arachnid Deployment
ARACHNID_BYTECODE="7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf3"
ARACHNID_CODEHASH="0x2fa86add0aed31f33a762c9d88e807c475bd51d0f52bd0955754b2608f7e4989"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

rm -rf "$HOMEDIR"

$BINARY config keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY config chain-id $CHAINID --home "$HOMEDIR"

echo "betray theory cargo way left cricket doll room donkey wire reunion fall left surprise hamster corn village happy bulb token artist twelve whisper expire" | $BINARY keys add alice --keyring-backend $KEYRING --home $HOMEDIR --recover
echo "toss sense candy point cost rookie jealous snow ankle electric sauce forward oblige tourist stairs horror grunt tenant afford master violin final genre reason" | $BINARY keys add bob --keyring-backend $KEYRING --home $HOMEDIR --recover
echo "offer feel open ancient relax habit field right evoke ball organ beauty" | $BINARY keys add test1 --recover  --keyring-backend $KEYRING --home "$HOMEDIR"
echo "olympic such citizen any bind small neutral hidden prefer pupil trash lemon" | $BINARY keys add test2 --recover  --keyring-backend $KEYRING --home "$HOMEDIR"
echo "cup hip eyebrow flock slogan filter gas tent angle purpose rose setup" | $BINARY keys add operator --recover --keyring-backend $KEYRING --home "$HOMEDIR"

$BINARY init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

jq '.app_state["feemarket"]["params"]["base_fee"]="7"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["bond_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["unbonding_time"]="1s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["inflation"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["compliance"]["operators"]=[{"operator":"swtr1ml2knanpk8sv94f8h9g8vaf9k3yyfva4fykyn9", "operator_type": 1}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# add arachnid deployment proxy
jq --arg BYTECODE $ARACHNID_BYTECODE '.app_state.evm.accounts += [{"address":"0x4e59b44847b379578588920cA78FbF26c0B4956C", "code": $BYTECODE, "storage": []}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg CODE_HASH $ARACHNID_CODEHASH '.app_state.auth.accounts += [{"@type": "/ethermint.types.v1.EthAccount", "base_account": {"account_number": "6", "address": "swtr1fevmgjz8kdu40pvgjgx20ralymqtf9tvcggehm", "pub_key": null, "sequence": "1" }, "code_hash": $CODE_HASH}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set initial issuer in genesis
jq --arg BYTECODE $COMPLIANCE_PROXY_BYTECODE '.app_state.evm.accounts += [{"address":"0x2Fc0B35E41a9a2eA248a275269Af1c8B3a061167", "code": $BYTECODE, "storage": []}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.compliance.issuerDetails += [{"address": "swtr19lqtxhjp4x3w5fy2yafxntcu3vaqvyt827e4ct", "details": {"creator": "swtr1ml2knanpk8sv94f8h9g8vaf9k3yyfva4fykyn9", "description": "d", "legalEntity": "e", "logo": "l", "name": "n", "url": "u"}}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.compliance.addressDetails += [{"address": "swtr19lqtxhjp4x3w5fy2yafxntcu3vaqvyt827e4ct", "details": {"is_revoked": false, "is_verified": true, "verifications": []}}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg CODE_HASH $COMPLIANCE_PROXY_CODEHASH '.app_state.auth.accounts += [{"@type": "/ethermint.types.v1.EthAccount", "base_account": {"account_number": "5", "address": "swtr19lqtxhjp4x3w5fy2yafxntcu3vaqvyt827e4ct", "pub_key": null, "sequence": "1" }, "code_hash": $CODE_HASH}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# disable node logs. remove this line if you need all logs
sed -i 's/info/error/g' "$CONFIG"

# expose ports
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
sed -i 's/127.0.0.1:8545/0.0.0.0:8545/g' "$APP_TOML"
sed -i 's/127.0.0.1:8546/0.0.0.0:8546/g' "$APP_TOML"

# enable prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"

# disable unsafe eth endpoints
sed -i 's/unsafe-eth-endpoints-enabled = true/unsafe-eth-endpoints-enabled = false/' "$APP_TOML"

# set min gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0aswtr"/' "$APP_TOML"

# Change proposal periods to pass within a reasonable time for local testing
sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$HOMEDIR"/config/genesis.json
sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$HOMEDIR"/config/genesis.json

# set custom pruning settings
sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "100"/g' "$APP_TOML"
sed -i.bak 's/pruning-interval = "0"/pruning-interval = "500"/g' "$APP_TOML"

# Allocate genesis accounts
$BINARY add-genesis-account alice 100000000swtr --keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY add-genesis-account bob 100000000swtr --keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY add-genesis-account test1 100000000swtr --keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY add-genesis-account test2 100000000swtr --keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY add-genesis-account operator 100000000swtr --keyring-backend $KEYRING --home "$HOMEDIR"

# Sign genesis transaction
$BINARY gentx alice 1000000000000000000000aswtr --keyring-backend $KEYRING --chain-id $CHAINID --home "$HOMEDIR"

# Collect genesis tx
$BINARY collect-gentxs --home "$HOMEDIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
$BINARY validate-genesis --home "$HOMEDIR"

# Initialize epoch keys for local testnet
$BINARY testnet init-testnet-enclave