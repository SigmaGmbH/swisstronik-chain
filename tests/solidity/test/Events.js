const { expect } = require('chai')
const { sendShieldedTransaction } = require("../test/testUtils")

it('Should emit events correctly', async () => {
    const [signer] = await ethers.getSigners()

    const EventsContract = await ethers.getContractFactory('EventTest')
    const eventInstance = await EventsContract.deploy()
    await eventInstance.deployed()

    const tx = await sendShieldedTransaction(
        signer,
        eventInstance.address,
        eventInstance.interface.encodeFunctionData("storeWithEvent", [888])
    )
    const receipt = await tx.wait()
    const logs = receipt.logs.map(log => eventInstance.interface.parseLog(log))

    expect(logs.some(log => log.name === 'ValueStored1' && log.args[0].toNumber() == 888)).to.be.true
    expect(logs.some(log => log.name === 'ValueStored2' && log.args[0] === 'TestMsg' && log.args[1].toNumber() == 888)).to.be.true
    expect(logs.some(log => log.name === 'ValueStored3' && log.args[0] === 'TestMsg' && log.args[1].toNumber() == 888 && log.args[2].toNumber() == 888)).to.be.true
})