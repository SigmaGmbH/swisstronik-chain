describe('Message calls', () => {
    let contract

    before(async () => {
        const TestContract = await ethers.getContractFactory('TestMessageCall')
        contract = await TestContract.deploy()
        await contract.deployed()
    })

    it('Contracts should be able to interact', async () => {
        const TEST_ITERATIONS = 30
        const tx = await contract.benchmarkMessageCall(TEST_ITERATIONS)
        await tx.wait()
    })
})