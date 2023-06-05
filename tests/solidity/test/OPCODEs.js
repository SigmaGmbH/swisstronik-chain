const { expect } = require('chai')
const { sendShieldedTransaction } = require("./testUtils")

describe('OPCODE test', () => {
    let contractInstance
    const provider = new ethers.providers.JsonRpcProvider('http://***REMOVED***:8545')
    const signerPrivateKey = '87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8'

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