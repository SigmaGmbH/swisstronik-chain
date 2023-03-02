const {expect} = require("chai");
const {ethers} = require("hardhat")

describe('Storage', () => {
    let contract

    before(async () => {
        const TestContract = await ethers.getContractFactory('Storage')
        contract = await TestContract.deploy()
        await contract.deployed()
    })

    it('Should be able to set value', async () => {
        const value = Math.floor(Math.random() * 10000)
        const tx = await contract.store(value)
        await tx.wait()

        const retrievedValue = await contract.retrieve()
        expect(retrievedValue).to.be.equal(value)
    })
})