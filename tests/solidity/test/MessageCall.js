describe('Contract tests', () => {
    describe('Message calls', () => {
        let contract

        before(async () => {
            const TestContract = await ethers.getContractFactory('TestMessageCall')
            contract = await TestContract.deploy()
            await contract.deployed()
        })

        it('Contracts should be able to interact', async () => {
            const TEST_ITERATIONS = 30
            await contract.benchmarkMessageCall(TEST_ITERATIONS)
        })
    })
})