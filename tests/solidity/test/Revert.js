const { expect } = require('chai')
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Revert / Error', () => {
    let revertContract, signer

    before(async () => {
        const [ethersSigner] = await ethers.getSigners()
        const RevertContract = await ethers.getContractFactory('TestRevert')
        revertContract = await RevertContract.deploy()
        signer = ethersSigner
        await revertContract.deployed()
    })

    it('testRevert: should revert if provided value < 10', async () => {

    })

    it('testRevert: should not revert if provided value >= 10', async () => {
        const tx = await sendShieldedTransaction(
            signer,
            revertContract.address,
            revertContract.interface.encodeFunctionData("testRevert", [10])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => revertContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Passed')).to.be.true
    })

    it('testError: should return error if provided value < 10', async () => {

    })

    it('testError: should not return error if provided value >= 10', async () => {
        const tx = await sendShieldedTransaction(
            signer,
            revertContract.address,
            revertContract.interface.encodeFunctionData("testError", [10])
        )
        const receipt = await tx.wait()

        const logs = receipt.logs.map(log => revertContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Passed')).to.be.true
    })
})
