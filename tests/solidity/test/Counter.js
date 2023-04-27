const {anyValue} = require("@nomicfoundation/hardhat-chai-matchers/withArgs");
const {expect} = require("chai");
const { getNodePublicKey, encryptECDH } = require("swisstronik.js")
const {ethers} = require("hardhat")
const {sendShieldedTransaction, sendShieldedQuery} = require("./testUtils")

describe('Counter', () => {
    let counterContract
    const signerPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'

    before(async () => {
        const Counter = await ethers.getContractFactory('Counter')
        counterContract = await Counter.deploy()
        await counterContract.deployed()
    })

    it('DEBUG', async () => {
        // There was a problem with default provider from ethers
        const signers = await ethers.getSigners()
        const provider = new ethers.providers.JsonRpcProvider('http://localhost:8535')
        const balance = await provider.getBalance(signers[0].address)
        console.log('balance: ', balance, 'for account: ', signers[0].address)
        const countBeforeResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const txResponse = await sendShieldedTransaction(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("add", [])
        )
        console.log('resp: ', txResponse)
    })

    // it('Should add', async () => {
    //     const countBefore = await counterContract.counter()
    //
    //     await expect(counterContract.add())
    //         .to.emit(counterContract, 'Added')
    //         .to.emit(counterContract, 'Changed')
    //
    //     const countAfter = await counterContract.counter()
    //     expect(countAfter).to.be.equal(countBefore + 1)
    // })
    //
    // it('Should subtract', async () => {
    //     const countBefore = await counterContract.counter()
    //
    //     await expect(counterContract.subtract())
    //         .to.emit(counterContract, 'Changed')
    //
    //     const countAfter = await counterContract.counter()
    //     expect(countAfter).to.be.equal(countBefore - 1)
    // })
    //
    // it('Should revert correctly', async () => {
    //     const tx = await counterContract.subtract()
    //     await expect(tx.wait()).to.be.rejected
    // })
})