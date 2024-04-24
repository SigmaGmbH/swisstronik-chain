const { expect } = require('chai')
const hre = require('hardhat')
const { sendShieldedTransaction } = require("./testUtils")

describe('Revert / Error', () => {
    let revertContract, signer

    before(async () => {
        const [ethersSigner] = await hre.ethers.getSigners()
        const RevertContract = await ethers.getContractFactory('TestRevert')
        revertContract = await RevertContract.deploy()
        signer = ethersSigner
        await revertContract.deployed()
    })

    it('testRevert: should revert if provided value < 10', async () => {     
        let reason = ""
        try {
            const tx = await sendShieldedTransaction(
                signer,
                revertContract.address,
                revertContract.interface.encodeFunctionData("testRevert", [5])
            )
            await tx.wait()
        } catch (e) {
            reason = e.reason
        }

        expect(reason).to.contain("Expected value >= 10")
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
})
