const { sendShieldedTransaction } = require("./testUtils")

describe('Message calls', () => {
    let contract
    const provider = new ethers.providers.JsonRpcProvider('http://localhost:8545')
    const signerPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'

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