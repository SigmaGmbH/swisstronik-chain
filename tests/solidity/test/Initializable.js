const { expect } = require('chai')
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Initializable', () => {
    let lifecycle
    const provider = new ethers.providers.JsonRpcProvider('http://***REMOVED***:8545')
    const senderPrivateKey = '87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8'

    beforeEach(async () => {
        const LifecycleMock = await ethers.getContractFactory('LifecycleMock')
        lifecycle = await LifecycleMock.deploy()
        await lifecycle.deployed()
    })

    it('is not initialized', async () => {
        const response = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            lifecycle.address,
            lifecycle.interface.encodeFunctionData("hasInitialized", [])
        );
        const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
        expect(initialized).to.be.false
    })

    it('is not petrified', async () => {
        const response = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            lifecycle.address,
            lifecycle.interface.encodeFunctionData("isPetrified", [])
        );
        const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
        expect(petrified).to.be.false
    })

    describe('> Initialized', () => {
        beforeEach(async () => {
            const tx = await sendShieldedTransaction(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("initializeMock", [])
            )
            await tx.wait()
        })

        it('is initialized', async () => {
            const response = await sendShieldedQuery(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("hasInitialized", [])
            );
            const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
            expect(initialized).to.be.true
        })

        it('is not petrified', async () => {
            const response = await sendShieldedQuery(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("isPetrified", [])
            );
            const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
            expect(petrified).to.be.false
        })

        it('cannot be re-initialized', async () => {
            let failed = false
            try {
                await sendShieldedTransaction(
                    provider,
                    receiverPrivateKey,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("initializeMock", [])
                )
            } catch {
                failed = true
            }
    
            expect(failed).to.be.true
        })

        it('cannot be petrified', async () => {
            let failed = false
            try {
                await sendShieldedTransaction(
                    provider,
                    receiverPrivateKey,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("petrifyMock", [])
                )
            } catch {
                failed = true
            }
    
            expect(failed).to.be.true
        })
    })

    describe('> Petrified', () => {
        beforeEach(async () => {
            const tx = await sendShieldedTransaction(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("petrifyMock", [])
            )
            await tx.wait()
        })

        it('is not initialized', async () => {
            const response = await sendShieldedQuery(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("hasInitialized", [])
            );
            const initialized = lifecycle.interface.decodeFunctionResult("hasInitialized", response)[0]
            expect(initialized).to.be.false
        })

        it('is petrified', async () => {
            const response = await sendShieldedQuery(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("isPetrified", [])
            );
            const petrified = lifecycle.interface.decodeFunctionResult("isPetrified", response)[0]
            expect(petrified).to.be.true
        })

        it('cannot be petrified again', async () => {
            let failed = false
            try {
                await sendShieldedTransaction(
                    provider,
                    receiverPrivateKey,
                    lifecycle.address,
                    lifecycle.interface.encodeFunctionData("petrifyMock", [])
                )
            } catch {
                failed = true
            }
    
            expect(failed).to.be.true
        })

        it('has initialization block in the future', async () => {
            const petrifiedBlockResponse = await sendShieldedQuery(
                provider,
                senderPrivateKey,
                lifecycle.address,
                lifecycle.interface.encodeFunctionData("getInitializationBlock", [])
            );
            const petrifiedBlock = lifecycle.interface.decodeFunctionResult("getInitializationBlock", petrifiedBlockResponse)[0]
            const blockNumber = await provider.getBlockNumber()
            expect(petrifiedBlock).to.be.greaterThan(blockNumber)
        })
    })
})
