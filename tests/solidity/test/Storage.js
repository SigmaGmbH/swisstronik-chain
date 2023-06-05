const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Storage', () => {
    let contract
    const provider = new ethers.providers.JsonRpcProvider('http://***REMOVED***:8545')
    const signerPrivateKey = '87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8'

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