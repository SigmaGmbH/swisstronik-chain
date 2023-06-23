const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Counter', () => {
    let counterContract

    before(async () => {
        const Counter = await ethers.getContractFactory('Counter')
        counterContract = await Counter.deploy()
        await counterContract.deployed()
    })

    it('Should add', async () => {
        const [signer] = await ethers.getSigners()

        const countBeforeResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const tx = await sendShieldedTransaction(
            signer,
            counterContract.address,
            counterContract.interface.encodeFunctionData("add", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Added')).to.be.true
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() + 1)
    })

    it('Should subtract', async () => {
        const [signer] = await ethers.getSigners()

        const countBeforeResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const tx = await sendShieldedTransaction(
            signer,
            counterContract.address,
            counterContract.interface.encodeFunctionData("subtract", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() - 1)
    })

    it('Should revert correctly', async () => {
        const [signer] = await ethers.getSigners()

        let failed = false
        try {
            const tx = await sendShieldedTransaction(
                signer,
                counterContract.address,
                counterContract.interface.encodeFunctionData("subtract", [])
            )
            await tx.wait()
        } catch (e) {
            failed = e.reason.indexOf('reverted') !== -1
        }

        expect(failed).to.be.true
    })
})