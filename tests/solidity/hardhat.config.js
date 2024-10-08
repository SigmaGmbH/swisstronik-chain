require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: {
        compilers: [
            {version: "0.8.24"},
        ],
        overrides: {
            "contracts/tokens/WETH9.sol": {
                version: "0.5.5"
            },
            "contracts/opcodes/TransientStorage.sol": {
                version: "0.8.27",
                settings: {
                    evmVersion: "cancun"
                }
            }
        }
    },
    networks: {
        tronik: {
            url: "http://localhost:8545",
            accounts: [
                "D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB",
                "DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55",
            ],
            chainId: 1291
        },
    },
    mocha: {
        timeout: 100000000
    },
};
