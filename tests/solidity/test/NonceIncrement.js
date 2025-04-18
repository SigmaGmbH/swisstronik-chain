const { expect } = require('chai')

it('Should send a transaction with EIP-1559 flag', async function () {
    const [sender, receiver] = await ethers.getSigners()

    const nonceBefore = await sender.getTransactionCount();

    const tx = await sender.sendTransaction({
        to: receiver.address,
        value: 1000000000,
        gasLimit: 21000,
        type: 2,
    })
    await tx.wait()
    expect(tx.type).to.be.equal(2)

    const nonceAfter = await sender.getTransactionCount();
    expect(nonceAfter).to.be.equal(nonceBefore + 1);
});