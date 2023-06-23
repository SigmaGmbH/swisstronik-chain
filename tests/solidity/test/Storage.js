const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Storage', () => {
    let contract

    before(async () => {
        const TestContract = await ethers.getContractFactory('Storage')
        contract = await TestContract.deploy()
        await contract.deployed()
    })

    it('Should be able to set value', async () => {
        const [signer] = await ethers.getSigners()
        const value = Math.floor(Math.random() * 10000)
        const tx = await sendShieldedTransaction(
            signer,
            contract.address,
            contract.interface.encodeFunctionData("store", [value])
        )
        await tx.wait()

        const retrievedValueResponse = await sendShieldedQuery(
            ethers.provider,
            contract.address,
            contract.interface.encodeFunctionData("retrieve", [])
        );
        const retrievedValue = contract.interface.decodeFunctionResult("retrieve", retrievedValueResponse)[0]

        expect(retrievedValue).to.be.equal(value)
    })
})