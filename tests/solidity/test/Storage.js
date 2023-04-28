const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Storage', () => {
    let contract
    const provider = new ethers.providers.JsonRpcProvider('http://localhost:8535')
    const signerPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'

    before(async () => {
        const TestContract = await ethers.getContractFactory('Storage')
        contract = await TestContract.deploy()
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