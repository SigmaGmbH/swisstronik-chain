const {expect} = require('chai')

it('Should emit events correctly', async () => {
    const EventsContract = await ethers.getContractFactory('EventTest')
    const eventInstance = await EventsContract.deploy()
    await eventInstance.deployed()

    await expect(eventInstance.storeWithEvent(888))
        .to.emit(eventInstance, 'ValueStored1').withArgs(888)
        .to.emit(eventInstance, 'ValueStored2').withArgs('TestMsg', 888)
        .to.emit(eventInstance, 'ValueStored3').withArgs('TestMsg', 888, 888)
})