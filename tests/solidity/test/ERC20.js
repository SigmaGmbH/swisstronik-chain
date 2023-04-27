const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('ERC20', () => {
    let tokenContract
    const provider = new ethers.providers.JsonRpcProvider('http://localhost:8535')
    const senderPrivateKey = '0xC516DC17D909EFBB64A0C4A9EE1720E10D47C1BF3590A257D86EEB5FFC644D43'
    const receiverPrivateKey = '0xD9B4808CEBBB85114D77FCB89CDE28FE5C422A9B42305E90FBE09F7B6D5A0C63'

    before(async () => {
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000)
        await tokenContract.deployed()
    })

    it('Should return correct name and symbol', async () => {
        const nameResponse = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("name", [])
        );
        const name = tokenContract.interface.decodeFunctionResult("name", nameResponse)[0]
        expect(name).to.be.equal('test token')

        const symbolResponse = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("symbol", [])
        );
        const symbol = tokenContract.interface.decodeFunctionResult("symbol", symbolResponse)[0]
        expect(symbol).to.be.equal('TT')
    })

    it('ERC20 transfer', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 100

        const senderBalanceBefore = await tokenContract.balanceOf(sender.address)
        const receiverBalanceBefore = await tokenContract.balanceOf(receiver.address)

        await expect(tokenContract.connect(sender).transfer(receiver.address, amountToTransfer))
            .to.emit(tokenContract, "Transfer")
            .withArgs(sender.address, receiver.address, amountToTransfer)

        const senderBalanceAfter = await tokenContract.balanceOf(sender.address)
        const receiverBalanceAfter = await tokenContract.balanceOf(receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    // it('ERC20 transferFrom', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 100

    //     await expect(tokenContract.connect(sender).approve(receiver.address, amountToTransfer))
    //         .to.emit(tokenContract, "Approval")
    //         .withArgs(sender.address, receiver.address, amountToTransfer)

    //     const senderBalanceBefore = await tokenContract.balanceOf(sender.address)
    //     const receiverBalanceBefore = await tokenContract.balanceOf(receiver.address)

    //     await expect(tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer))
    //         .to.emit(tokenContract, "Transfer")
    //         .withArgs(sender.address, receiver.address, amountToTransfer)

    //     const senderBalanceAfter = await tokenContract.balanceOf(sender.address)
    //     const receiverBalanceAfter = await tokenContract.balanceOf(receiver.address)

    //     expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
    //     expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    // })

    // it('Cannot exceed balance during transfer', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 1000000000000

    //     const tx = await tokenContract.connect(sender).transfer(receiver.address, amountToTransfer)
    //     await expect(tx.wait()).to.be.rejected
    // })

    // it('Cannot transfer more than approved', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 1000000000000

    //     const tx = await tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer)
    //     await expect(tx.wait()).to.be.rejected
    // })
})