require('dotenv').config()
const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery, getProvider } = require("../test/testUtils")

const getTokenBalance = async (provider, privateKey, contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        provider,
        privateKey,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

describe('ERC20', () => {
    let tokenContract
    const provider = getProvider()
    const senderPrivateKey = process.env.FIRST_PRIVATE_KEY
    const receiverPrivateKey = process.env.SECOND_PRIVATE_KEY

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

        const senderBalanceBefore = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
        const receiverBalanceBefore = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

        const tx = await sendShieldedTransaction(
            provider,
            senderPrivateKey,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("transfer", [receiver.address, amountToTransfer])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceAfter = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
        const receiverBalanceAfter = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    it('ERC20 transferFrom', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 100

        const approveTx = await sendShieldedTransaction(
            provider,
            senderPrivateKey,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("approve", [receiver.address, amountToTransfer])
        )
        const approveTxReceipt = await approveTx.wait()
        const approveTxLogs = approveTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(approveTxLogs.some(log => log.name === 'Approval' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceBefore = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
        const receiverBalanceBefore = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

        const transferTx = await sendShieldedTransaction(
            provider,
            receiverPrivateKey,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("transferFrom", [sender.address, receiver.address, amountToTransfer])
        )
        const transferTxReceipt = await transferTx.wait()
        const transferTxLogs = transferTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
        expect(transferTxLogs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

        const senderBalanceAfter = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
        const receiverBalanceAfter = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

        expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
        expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    })

    it('Cannot exceed balance during transfer', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 1000000000000

        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                senderPrivateKey,
                tokenContract.address,
                tokenContract.interface.encodeFunctionData("transfer", [receiver.address, amountToTransfer])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })

    it('Cannot transfer more than approved', async () => {
        const [sender, receiver] = await ethers.getSigners()
        const amountToTransfer = 1000000000000

        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                receiverPrivateKey,
                tokenContract.address,
                tokenContract.interface.encodeFunctionData("transferFrom", [sender.address, receiver.address, amountToTransfer])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })
})