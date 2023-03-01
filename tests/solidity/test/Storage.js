const {expect} = require("chai");

describe('Storage', () => {
    let contract

    before(async () => {
        const TestContract = await ethers.getContractFactory('Storage')
        contract = await TestContract.deploy()
    })

    it('Should be able to set value', async () => {
        const value = Math.floor(Math.random() * 10000)
        await contract.store(value)

        const retrievedValue = await contract.retrieve()
        expect(retrievedValue).to.be.equal(value)
    })
})