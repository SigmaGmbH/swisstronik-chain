const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Counter', () => {
    let counterContract
    const provider = new ethers.providers.JsonRpcProvider('http://localhost:8535')
    const signerPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'

    before(async () => {
        const Counter = await ethers.getContractFactory('Counter')
        counterContract = await Counter.deploy()
        await counterContract.deployed()
    })

    it('Should add', async () => {
        const countBeforeResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const tx = await sendShieldedTransaction(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("add", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Added')).to.be.true
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() + 1)
    })

    it('Should subtract', async () => {
        const countBeforeResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const tx = await sendShieldedTransaction(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("subtract", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() - 1)
    })

    it('Should revert correctly', async () => {
        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                signerPrivateKey,
                counterContract.address,
                counterContract.interface.encodeFunctionData("subtract", [])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })
})