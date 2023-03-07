const {expect} = require('chai')

it('Should send a transaction with EIP-1559 flag', async function () {
    const [sender, receiver] = await ethers.getSigners()
    const tx = await sender.sendTransaction({
        to: receiver.address,
        value: 100,
        gasLimit: 21000,
        type: 2
    })
    await tx.wait()
    expect(tx.type).to.be.equal(2)
});