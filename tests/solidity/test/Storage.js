require('dotenv').config()
const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery, getProvider } = require("./testUtils")

describe('Storage', () => {
    let contract
    const provider = getProvider()
    const signerPrivateKey = process.env.FIRST_PRIVATE_KEY

    before(async () => {
        const TestContract = await ethers.getContractFactory('Storage')
        contract = await TestContract.deploy({gasLimit: 1000000})
        await contract.deployed()
    })

    it('Should be able to set value', async () => {
        const value = Math.floor(Math.random() * 10000)
        const tx = await sendShieldedTransaction(
            provider,
            signerPrivateKey,
            contract.address,
            contract.interface.encodeFunctionData("store", [value])
        )
        await tx.wait()

        const retrievedValueResponse = await sendShieldedQuery(
            provider,
            signerPrivateKey,
            contract.address,
            contract.interface.encodeFunctionData("retrieve", [])
        );
        const retrievedValue = contract.interface.decodeFunctionResult("retrieve", retrievedValueResponse)[0]

        expect(retrievedValue).to.be.equal(value)
    })
})