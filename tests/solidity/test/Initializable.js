const { expect } = require('chai')
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Initializable', () => {
    let lifecycle

    beforeEach(async () => {
        const LifecycleMock = await ethers.getContractFactory('LifecycleMock')
        lifecycle = await LifecycleMock.deploy()
        await lifecycle.deployed()
    })

    it('is not initialized', async () => {
        const response = await sendShieldedQuery(
            ethers.provider,
            lifecycle.address,
            lifecycle.interface.encodeFunctionData("hasInitialized", [])
        );
        const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
        expect(initialized).to.be.false
    })

    it('is not petrified', async () => {
        const response = await sendShieldedQuery(
            ethers.provider,
            lifecycle.address,
            lifecycle.interface.encodeFunctionData("isPetrified", [])
        );
        const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
        expect(petrified).to.be.false
    })

    describe('> Initialized', () => {
        beforeEach(async () => {
            const [signer] = await ethers.getSigners()
            const tx = await sendShieldedTransaction(
                signer,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("initializeMock", [])
            )
            await tx.wait()
        })

        it('is initialized', async () => {
            const response = await sendShieldedQuery(
                ethers.provider,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("hasInitialized", [])
            );
            const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
            expect(initialized).to.be.true
        })

        it('is not petrified', async () => {
            const response = await sendShieldedQuery(
                ethers.provider,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("isPetrified", [])
            );
            const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
            expect(petrified).to.be.false
        })

        it('cannot be re-initialized', async () => {
            const [signer] = await ethers.getSigners()
            let failed = false
            try {
                await sendShieldedTransaction(
                    signer,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("initializeMock", [])
                )
            } catch (e) {
                failed = e.reason.indexOf('reverted') !== -1
            }
    
            expect(failed).to.be.true
        })

        it('cannot be petrified', async () => {
            const [signer] = await ethers.getSigners()
            let failed = false
            try {
                await sendShieldedTransaction(
                    signer,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("petrifyMock", [])
                )
            } catch (e) {
                failed = e.reason.indexOf('reverted') !== -1
            }
    
            expect(failed).to.be.true
        })
    })

    describe('> Petrified', () => {
        beforeEach(async () => {
            const [signer] = await ethers.getSigners()
            const tx = await sendShieldedTransaction(
                signer,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("petrifyMock", [])
            )
            await tx.wait()
        })

        it('is not initialized', async () => {
            const response = await sendShieldedQuery(
                ethers.provider,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("hasInitialized", [])
            );
            const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
            expect(initialized).to.be.false
        })

        it('is petrified', async () => {
            const response = await sendShieldedQuery(
                ethers.provider,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("isPetrified", []))
            const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
            expect(petrified).to.be.true
        })

        it('cannot be petrified again', async () => {
            const [signer] = await ethers.getSigners()
            let failed = false
            try {
                await sendShieldedTransaction(
                    signer,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("petrifyMock", [])
                )
            } catch (e) {
                failed = e.reason.indexOf('reverted') !== -1
            }
    
            expect(failed).to.be.true
        })

        it('has initialization block in the future', async () => {
            const petrifiedBlockResponse = await sendShieldedQuery(
                ethers.provider,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("getInitializationBlock", [])
            );
            const petrifiedBlock = lifecycle.interface.decodeFunctionResult("getInitializationBlock", petrifiedBlockResponse)[0]
            const blockNumber = await ethers.provider.getBlockNumber()
            expect(petrifiedBlock).to.be.greaterThan(blockNumber)
        })
    })
})
