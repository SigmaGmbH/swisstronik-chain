const { expect } = require('chai')
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

it('Should revert', async () => {
    const [signer] = await ethers.getSigners()
    const RevertContract = await ethers.getContractFactory('TestRevert')
    const revertContract = await RevertContract.deploy()
    await revertContract.deployed()

    const trySetTx = await sendShieldedTransaction(
        signer,
        revertContract.address,
        revertContract.interface.encodeFunctionData("try_set", [10])
    )
    await trySetTx.wait()

    const queryAResponse = await sendShieldedQuery(
        ethers.provider,
        revertContract.address,
        revertContract.interface.encodeFunctionData("query_a", [])
    );
    const queryAResult = revertContract.interface.decodeFunctionResult("query_a", queryAResponse)[0]
    expect(queryAResult).to.be.equal(0)

    const queryBResponse = await sendShieldedQuery(
        ethers.provider,
        revertContract.address,
        revertContract.interface.encodeFunctionData("query_b", [])
    );
    const queryBResult = revertContract.interface.decodeFunctionResult("query_b", queryBResponse)[0]
    expect(queryBResult).to.be.equal(10)

    const queryCResponse = await sendShieldedQuery(
        ethers.provider,
        revertContract.address,
        revertContract.interface.encodeFunctionData("query_c", [])
    );
    const queryCResult = revertContract.interface.decodeFunctionResult("query_c", queryCResponse)[0]
    expect(queryCResult).to.be.equal(10)

    const setTx = await sendShieldedTransaction(
        signer,
        revertContract.address,
        revertContract.interface.encodeFunctionData("set", [10])
    )
    await setTx.wait()

    const response = await sendShieldedQuery(
        ethers.provider,
        revertContract.address,
        revertContract.interface.encodeFunctionData("query_a", [])
    );
    const result = revertContract.interface.decodeFunctionResult("query_a", response)[0]
    expect(result).to.be.equal(10)
})
