const {expect} = require("chai");

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
        await expect(counterContract.subtract())
            .to.be.revertedWith("COUNTER_TOO_LOW")
    })
})