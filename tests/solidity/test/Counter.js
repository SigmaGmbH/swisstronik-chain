const { anyValue } = require("@nomicfoundation/hardhat-chai-matchers/withArgs");
const { expect } = require("chai");

describe('Contract tests', () => {
    describe('Counter', () => {
        let counterContract

        before(async () => {
            const Counter = await ethers.getContractFactory('Counter')
            counterContract = await Counter.deploy()
        })

        it('Should add', async () => {
            const countBefore = await counterContract.counter()
            await counterContract.add()
            const countAfter = await counterContract.counter()
            expect(countAfter).to.be.equal(countBefore + 1)
        })
    })
})