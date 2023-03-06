const {expect} = require('chai')

it('Should revert', async () => {
    const RevertContract = await ethers.getContractFactory('TestRevert')
    const revertContract = await RevertContract.deploy()
    await revertContract.deployed()

    const trySetTx = await revertContract.try_set(10)
    await trySetTx.wait()

    expect(await revertContract.query_a()).to.be.equal(0)
    expect(await revertContract.query_b()).to.be.equal(10)
    expect(await revertContract.query_c()).to.be.equal(10)

    const setTx = await revertContract.set(10)
    await setTx.wait()

    expect(await revertContract.query_a()).to.be.equal(10)
})
