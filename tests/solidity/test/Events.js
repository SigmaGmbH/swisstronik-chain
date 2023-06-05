const { expect } = require('chai')
const { sendShieldedTransaction } = require("../test/testUtils")

it('Should emit events correctly', async () => {
    const provider = new ethers.providers.JsonRpcProvider('http://***REMOVED***:8545')
    const signerPrivateKey = '87D17E1D032E65CA33435C35144457EE1F12B8B4E706C6795728E998780AFCD8'

    const EventsContract = await ethers.getContractFactory('EventTest')
    const eventInstance = await EventsContract.deploy()
    await eventInstance.deployed()

    const tx = await sendShieldedTransaction(
        provider,
        signerPrivateKey,
        eventInstance.address,
        eventInstance.interface.encodeFunctionData("storeWithEvent", [888])
    )
    const receipt = await tx.wait()
    const logs = receipt.logs.map(log => eventInstance.interface.parseLog(log))

    expect(logs.some(log => log.name === 'ValueStored1' && log.args[0].toNumber() == 888)).to.be.true
    expect(logs.some(log => log.name === 'ValueStored2' && log.args[0] === 'TestMsg' && log.args[1].toNumber() == 888)).to.be.true
    expect(logs.some(log => log.name === 'ValueStored3' && log.args[0] === 'TestMsg' && log.args[1].toNumber() == 888 && log.args[2].toNumber() == 888)).to.be.true
})