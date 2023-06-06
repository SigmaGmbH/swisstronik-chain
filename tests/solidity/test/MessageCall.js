require('dotenv').config()
const { sendShieldedTransaction, getProvider } = require("./testUtils")

describe('Message calls', () => {
    let contract
    const provider = getProvider()
    const signerPrivateKey = process.env.FIRST_PRIVATE_KEY

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