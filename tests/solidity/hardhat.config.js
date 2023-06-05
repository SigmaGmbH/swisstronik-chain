require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    solidity: "0.8.17",
    networks: {
        "tronik": {
            url: "http://***REMOVED***:8545",
            accounts: [
                "87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8",
                "247991D4707FE6C67756C90BD324EE4508E12DD7ED0DEF003281345781605204",
            ],
            gas: 3_000_000
        },
        "local": {
            url: "http://***REMOVED***:8545",
            accounts: [
                "87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8",
                "247991D4707FE6C67756C90BD324EE4508E12DD7ED0DEF003281345781605204",
            ],
            gas: 3_000_000
        },
    },
    mocha: {
        timeout: 100000000
      },
};
