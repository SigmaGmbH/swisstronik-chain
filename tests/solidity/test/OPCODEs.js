const { expect } = require('chai')
const { sendShieldedTransaction } = require("./testUtils")

describe('OPCODE test', () => {
    let contractInstance
    const provider = new ethers.providers.JsonRpcProvider('http://localhost:8535')
    const signerPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'

    beforeEach(async () => {
        const OpcodesContract = await ethers.getContractFactory('OpCodes')
        contractInstance = await OpcodesContract.deploy()
        await contractInstance.deployed()
    })

    it('Should throw invalid op code', async () => {
        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                signerPrivateKey,
                contractInstance.address,
                contractInstance.interface.encodeFunctionData("test_invalid", [])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })

    it('Should revert', async () => {
        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                signerPrivateKey,
                contractInstance.address,
                contractInstance.interface.encodeFunctionData("test_revert", [])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })
});