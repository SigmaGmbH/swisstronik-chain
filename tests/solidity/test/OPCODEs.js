const { expect } = require('chai')
const { sendShieldedTransaction } = require("./testUtils")

describe('OPCODE test', () => {
    let contractInstance

    beforeEach(async () => {
        const OpcodesContract = await ethers.getContractFactory('OpCodes')
        contractInstance = await OpcodesContract.deploy()
        await contractInstance.deployed()
    })

    it('Should throw invalid op code', async () => {
        const [signer] = await ethers.getSigners()
        let failed = false
        try {
            const tx = await sendShieldedTransaction(
                signer,
                contractInstance.address,
                contractInstance.interface.encodeFunctionData("test_invalid", [])
            )
            await tx.wait()
        } catch (e) {
            failed = e.reason.indexOf('reverted') !== -1
        }

        expect(failed).to.be.true
    })

    it('Should revert', async () => {
        const [signer] = await ethers.getSigners()
        let failed = false
        try {
            const tx = await sendShieldedTransaction(
                signer,
                contractInstance.address,
                contractInstance.interface.encodeFunctionData("test_revert", [])
            )
            await tx.wait()
        } catch (e) {
            failed = e.reason.indexOf('reverted') !== -1
        }

        expect(failed).to.be.true
    })
});