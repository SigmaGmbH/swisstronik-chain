const {ethers} = require('hardhat')
const {expect} = require('chai')

describe('Initializable', () => {
    let lifecycle

    beforeEach(async () => {
        const LifecycleMock = await ethers.getContractFactory('LifecycleMock')
        lifecycle = await LifecycleMock.deploy()
        await lifecycle.deployed()
    })

    it('is not initialized', async () => {
        expect(await lifecycle.hasInitialized()).to.be.false
    })

    it('is not petrified', async () => {
        expect(await lifecycle.isPetrified()).to.be.false
    })

    describe('> Initialized', () => {
        beforeEach(async () => {
            const tx = await lifecycle.initializeMock()
            await tx.wait()
        })

        it('is initialized', async () => {
            expect(await lifecycle.hasInitialized()).to.be.true
        })

        it('is not petrified', async () => {
            expect(await lifecycle.isPetrified()).to.be.false
        })

        it('cannot be re-initialized', async () => {
            const tx = await lifecycle.initializeMock()
            await expect(tx.wait()).to.be.rejected
        })

        it('cannot be petrified', async () => {
            const tx = await lifecycle.petrifyMock()
            await expect(tx.wait()).to.be.rejected
        })
    })

    describe('> Petrified', () => {
        beforeEach(async () => {
            const tx = await lifecycle.petrifyMock()
            await tx.wait()
        })

        it('is not initialized', async () => {
            expect(await lifecycle.hasInitialized()).to.be.false
        })

        it('is petrified', async () => {
            expect(await lifecycle.isPetrified()).to.be.true
        })

        it('cannot be petrified again', async () => {
            const tx = await lifecycle.petrifyMock()
            await expect(tx.wait()).to.be.rejected
        })

        it('has initialization block in the future', async () => {
            const petrifiedBlock = await lifecycle.getInitializationBlock()
            const blockNumber = await ethers.getDefaultProvider().getBlockNumber()
            expect(petrifiedBlock).to.be.greaterThan(blockNumber)
        })
    })
})
