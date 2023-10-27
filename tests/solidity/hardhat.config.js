require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        tronik: {
            url: "http://localhost:8545",
            accounts: [
                "361CB3A8B12487FA469762D3EEC8AE79CFB6D6FF119D4DA28FACAB8A15184B16",
                "BF54A27255AAF74AA2B85BFB0C30F899D03875EEBFC54373E1381F34E63EF2E5",
            ],
            chainId: 1291
        },
    },
    mocha: {
        timeout: 100000000
    },
};
