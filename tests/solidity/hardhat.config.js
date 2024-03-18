require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        tronik: {
            url: "http://localhost:8545",
            accounts: [
                "DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55",
                "13B2DE5CFDF24796472F572B87D0732406B323F70A38FA244FE0A86736B77DFA",
            ],
            chainId: 1291
        },
    },
    mocha: {
        timeout: 100000000
    },
};
