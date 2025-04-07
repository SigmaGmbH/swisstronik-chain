const { expect } = require("chai")
const { ethers } = require("hardhat")

const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const sender = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)
const receiver = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", provider)

describe('ERC20 Unencrypted', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000)
        await tokenContract.deployed()
    })

    it('Should return correct name and symbol', async () => {
        const name = await tokenContract.connect(provider).name();
        expect(name).to.be.equal('test token')

        const symbol = await tokenContract.connect(provider).symbol();
        expect(symbol).to.be.equal('TT')
    })

    it('ERC20 transfer', async () => {
        const amountToTransfer = 100

        const senderBalanceBefore = await tokenContract.connect(provider).balanceOf(sender.address)
        const receiverBalanceBefore = await tokenContract.connect(provider).balanceOf(receiver.address)

        const tx = await tokenContract.connect(sender).transfer(receiver.address, amountToTransfer)
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceAfter = await tokenContract.connect(provider).balanceOf(sender.address)
        const receiverBalanceAfter = await tokenContract.connect(provider).balanceOf(receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    it('ERC20 transferFrom', async () => {
        const amountToTransfer = 100

        const approveTx = await tokenContract.connect(sender).approve(receiver.address, amountToTransfer)
        const approveTxReceipt = await approveTx.wait()
        const approveTxLogs = approveTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(approveTxLogs.some(log => log.name === 'Approval' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceBefore = await tokenContract.connect(provider).balanceOf(sender.address)
        const receiverBalanceBefore = await tokenContract.connect(provider).balanceOf(receiver.address)

        const transferTx = await tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer)
        const transferTxReceipt = await transferTx.wait()
        const transferTxLogs = transferTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(transferTxLogs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceAfter = await tokenContract.connect(provider).balanceOf(sender.address)
        const receiverBalanceAfter = await tokenContract.connect(provider).balanceOf(receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    it('Cannot exceed balance during transfer', async () => {
        const amountToTransfer = 1000000000000

        let failed = false
        try {
            const tx = await tokenContract.connect(receiver).transfer(receiver.address, amountToTransfer)
            await tx.wait()
        } catch (e) {
            failed = true
        }

        expect(failed).to.be.true
    })

    it('Cannot transfer more than approved', async () => {
        const amountToTransfer = 1000000000000

        let failed = false
        try {
            const tx = await tokenContract.connect(receiver).transferFrom(sender.address, receiver.address, amountToTransfer)
            await tx.wait()
        } catch (e) {
            failed = true
        }

        expect(failed).to.be.true
    })
})