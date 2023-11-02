require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        tronik: {
            url: "http://localhost:8545",
            accounts: [
                "F8FBAD01F31AF55D55A6967AAEA8ABB42E301F085895ADA0FB0C3BC2A5BA9371",
                "13B2DE5CFDF24796472F572B87D0732406B323F70A38FA244FE0A86736B77DFA",
            ],
            chainId: 1291
        },
    },
    mocha: {
        timeout: 100000000
    },
};
