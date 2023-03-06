const {expect} = require('chai')

describe('OPCODE test', () => {
    let contractInstance

    beforeEach(async () => {
        const OpcodesContract = await ethers.getContractFactory('OpCodes')
        contractInstance = await OpcodesContract.deploy()
        await contractInstance.deployed()
    })

    it('Should throw invalid op code', async () => {
        const tx = await contractInstance.test_invalid()
        await expect(tx.wait()).to.be.rejected
    })

    it('Should revert', async () => {
        const tx = await contractInstance.test_revert()
        await expect(tx.wait()).to.be.rejected
    })
});