const { sendShieldedTransaction } = require("./testUtils")

describe('Message calls', () => {
    let contract

    before(async () => {
        const TestContract = await ethers.getContractFactory('TestMessageCall')
        contract = await TestContract.deploy()
        await contract.deployed()
    })

    it('Contracts should be able to interact', async () => {
        const [signer] = await ethers.getSigners()
        const TEST_ITERATIONS = 30
        const tx = await sendShieldedTransaction(
            signer,
            contract.address,
            contract.interface.encodeFunctionData("benchmarkMessageCall", [TEST_ITERATIONS])
        )
        await tx.wait()
    })
})