const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery } = require("../test/testUtils")

const getTokenBalance = async (contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        ethers.provider,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

describe('ERC20', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000)
        await tokenContract.deployed()
    })

    it('Should return correct name and symbol', async () => {
        const nameResponse = await sendShieldedQuery(
            ethers.provider,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("name", [])
        );
        const name = tokenContract.interface.decodeFunctionResult("name", nameResponse)[0]
        expect(name).to.be.equal('test token')

        const symbolResponse = await sendShieldedQuery(
            ethers.provider,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("symbol", [])
        );
        const symbol = tokenContract.interface.decodeFunctionResult("symbol", symbolResponse)[0]
        expect(symbol).to.be.equal('TT')
    })

    // it('ERC20 transfer', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 100
    //
    //     const senderBalanceBefore = await getTokenBalance(tokenContract, sender.address)
    //     const receiverBalanceBefore = await getTokenBalance(tokenContract, receiver.address)
    //
    //     const tx = await sendShieldedTransaction(
    //         sender,
    //         tokenContract.address,
    //         tokenContract.interface.encodeFunctionData("transfer", [receiver.address, amountToTransfer])
    //     )
    //     const receipt = await tx.wait()
    //     const logs = receipt.logs.map(log => tokenContract.interface.parseLog(log))
    //     expect(logs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true
    //
    //     const senderBalanceAfter = await getTokenBalance(tokenContract, sender.address)
    //     const receiverBalanceAfter = await getTokenBalance(tokenContract, receiver.address)
    //
    //     expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
    //     expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    // })
    //
    // it('ERC20 transferFrom', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 100
    //
    //     const approveTx = await sendShieldedTransaction(
    //         sender,
    //         tokenContract.address,
    //         tokenContract.interface.encodeFunctionData("approve", [receiver.address, amountToTransfer])
    //     )
    //     const approveTxReceipt = await approveTx.wait()
    //     const approveTxLogs = approveTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
    //     expect(approveTxLogs.some(log => log.name === 'Approval' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true
    //
    //     const senderBalanceBefore = await getTokenBalance(tokenContract, sender.address)
    //     const receiverBalanceBefore = await getTokenBalance(tokenContract, receiver.address)
    //
    //     const transferTx = await sendShieldedTransaction(
    //         receiver,
    //         tokenContract.address,
    //         tokenContract.interface.encodeFunctionData("transferFrom", [sender.address, receiver.address, amountToTransfer])
    //     )
    //     const transferTxReceipt = await transferTx.wait()
    //     const transferTxLogs = transferTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
    //     expect(transferTxLogs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true
    //
    //     const senderBalanceAfter = await getTokenBalance(tokenContract, sender.address)
    //     const receiverBalanceAfter = await getTokenBalance(tokenContract, receiver.address)
    //
    //     expect(senderBalanceAfter.toNumber()).to.be.equal(senderBalanceBefore.toNumber() - amountToTransfer)
    //     expect(receiverBalanceAfter.toNumber()).to.be.equal(receiverBalanceBefore.toNumber() + amountToTransfer)
    // })

    it('Cannot exceed balance during transfer', async () => {
        const [_, receiver] = await ethers.getSigners()
        const amountToTransfer = 1000000000000

        let failed = false
        try {
            const tx = await sendShieldedTransaction(
                receiver,
                tokenContract.address,
                tokenContract.interface.encodeFunctionData("transfer", [receiver.address, amountToTransfer])
            )
            await tx.wait()
        } catch (e) {
            failed = e.reason.indexOf('reverted') !== -1
        }

        expect(failed).to.be.true
    })

    // it('Cannot transfer more than approved', async () => {
    //     const [sender, receiver] = await ethers.getSigners()
    //     const amountToTransfer = 1000000000000
    //
    //     let failed = false
    //     try {
    //         const tx = await sendShieldedTransaction(
    //             receiver,
    //             tokenContract.address,
    //             tokenContract.interface.encodeFunctionData("transferFrom", [sender.address, receiver.address, amountToTransfer])
    //         )
    //         await tx.wait()
    //     } catch (e) {
    //         failed = e.reason.indexOf('reverted') !== -1
    //     }
    //
    //     expect(failed).to.be.true
    // })
})