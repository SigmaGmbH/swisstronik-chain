const { sendShieldedTransaction } = require("./testUtils")

describe('Message calls', () => {
    let contract
    const provider = new ethers.providers.JsonRpcProvider('http://***REMOVED***:8545')
    const signerPrivateKey = '87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8'

    before(async () => {
        const TestContract = await ethers.getContractFactory('TestMessageCall')
        contract = await TestContract.deploy()
        await contract.deployed()
    })

    it('Contracts should be able to interact', async () => {
        const TEST_ITERATIONS = 30
        const tx = await sendShieldedTransaction(
            provider,
            signerPrivateKey,
            contract.address,
            contract.interface.encodeFunctionData("benchmarkMessageCall", [TEST_ITERATIONS])
        )
        await tx.wait()
    })
})