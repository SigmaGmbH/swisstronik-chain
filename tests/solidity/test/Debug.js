const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Debug', () => {
    // it('DEBUG', async () => {
    //     const [signer] = await ethers.getSigners()
    //     const nonce = await signer.provider.getTransactionCount(signer.address)
    //
    //     console.log('Caller: ', signer.address)
    //     console.log('Caller nonce: ', nonce)
    //     const calculatedAddr = ethers.utils.getContractAddress({from: signer.address, nonce: nonce})
    //     console.log('Exp address: ', calculatedAddr) // 0xa079a9f7d09fb132e95aa627c2ff414a43787c69
    //
    //     // 1 - 0x56f078d59950b21251e32fd659c89f80cf669fcc js: 0x30252aFE8C1683fD184C99a3c44Aa5d547d59dd4
    //     // 0 -   js: 0x54014e46667907922c58Ac224dFb0848f43EA800
    //     // 2 -   js: 0x202B7E8CaF217a1bc54cBFcA986d17B268cdd9d7
    //     // 3 -   js: 0x56f078d59950b21251E32Fd659c89f80Cf669fcc
    //
    //     //
    //     const Counter = await ethers.getContractFactory('Counter')
    //     const counterContract = await Counter.deploy({gasLimit: 1_000_000})
    //     await counterContract.deployed()
    //
    //     console.log('Counter deployed to: ', counterContract.address)
    //
    //     const contractCode = await signer.provider.getCode(counterContract.address)
    //     console.log('Contract code: ', contractCode)
    // })
    it('Calculate contract address', async () => {
            const [signer] = await ethers.getSigners()
            const nonce = await signer.provider.getTransactionCount(signer.address)

            console.log('Caller: ', signer.address)
            console.log('Caller nonce: ', nonce)
            const calculatedAddr = ethers.utils.getContractAddress({from: signer.address, nonce})
            console.log('Exp address: ', calculatedAddr)

        // test 0x2Fc0B35E41a9a2eA248a275269Af1c8B3a061167
    })
})