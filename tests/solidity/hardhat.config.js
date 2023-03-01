require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        "tronik": {
            url: "http://127.0.0.1:8535",
            accounts: [
                "0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43",
                "831052AB296006AA0366652BC01C2CA8E46621555E9F45FA353C80523225F756",
            ]
        },
        "local": {
            url: "http://127.0.0.1:8545",
            mnemonic: "gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat",
        }
    }
};
