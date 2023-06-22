require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        tronik: {
            url: "http://localhost:8545",
            accounts: [
                "C516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43",
                "B871BDE7A13FA4CE6B3933B6AF6A854AAA01E954C71B6FC8D3DAF7FD67DE9910",
            ],
            chainId: 1291
        },
    },
    mocha: {
        timeout: 100000000
    },
};
