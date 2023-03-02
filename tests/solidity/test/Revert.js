const {expect} = require("chai");
const {ethers} = require("hardhat")

// This is a test only for debug purpose.
// It will be removed, when issue with incorrect revert messages will be solved
describe('Counter', () => {
    let counterContract

    before(async () => {
        const Counter = await ethers.getContractFactory('Counter')
        counterContract = await Counter.deploy()
        await counterContract.deployed()
    })

    it('Should revert correctly', async () => {
        const tx = await counterContract.subtract()
        await expect(tx.wait()).to.be.rejected
    })
})