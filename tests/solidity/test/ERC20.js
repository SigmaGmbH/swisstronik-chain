const {expect} = require("chai")
const {ethers} = require("hardhat")

describe('ERC20', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000)
        await tokenContract.deployed()
    })

    it('Should return correct name and symbol', async () => {
        expect(await tokenContract.name()).to.be.equal('test token')
        expect(await tokenContract.symbol()).to.be.equal('TT')
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

    it('ERC20 transferFrom', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 100

        await expect(tokenContract.connect(sender).approve(receiver.address, amountToTransfer))
            .to.emit(tokenContract, "Approval")
            .withArgs(sender.address, receiver.address, amountToTransfer)

        const senderBalanceBefore = await tokenContract.balanceOf(sender.address)
        const receiverBalanceBefore = await tokenContract.balanceOf(receiver.address)

        await expect(tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer))
            .to.emit(tokenContract, "Transfer")
            .withArgs(sender.address, receiver.address, amountToTransfer)

        const senderBalanceAfter = await tokenContract.balanceOf(sender.address)
        const receiverBalanceAfter = await tokenContract.balanceOf(receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    it('Cannot exceed balance during transfer', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 1000000000000

        const tx = await tokenContract.connect(sender).transfer(receiver.address, amountToTransfer)
        await expect(tx.wait()).to.be.rejected
    })

    it('Cannot transfer more than approved', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 1000000000000

        const tx = await tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer)
        await expect(tx.wait()).to.be.rejected
    })
})