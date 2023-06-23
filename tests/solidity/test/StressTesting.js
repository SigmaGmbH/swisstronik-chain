const { expect } = require("chai")
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

const NUM_TESTING_ACCOUNTS = 50;

const getTokenBalance = async (contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        ethers.provider,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

const transferERC20Token = async (sender, receiverAddress, tokenContract, amountToTransfer) => {
    const senderBalanceBefore = await getTokenBalance(tokenContract, sender.address)
    const receiverBalanceBefore = await getTokenBalance(tokenContract, receiverAddress)

    const tx = await sendShieldedTransaction(
        sender,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("transfer", [receiverAddress, amountToTransfer])
    )
    await tx.wait()
    const senderBalanceAfter = await getTokenBalance(tokenContract, sender.address)
    const receiverBalanceAfter = await getTokenBalance(tokenContract, receiverAddress)

    return [senderBalanceBefore.toNumber(), senderBalanceAfter.toNumber(), receiverBalanceBefore.toNumber(), receiverBalanceAfter.toNumber()]
}

const transferFromERC20Token  = async (sender, receiver, tokenContract, amountToTransfer) => {
    const approveTx = await sendShieldedTransaction(
        sender,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("approve", [receiver.address, amountToTransfer])
    )
    const approveTxReceipt = await approveTx.wait()
    const approveTxLogs = approveTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
    expect(approveTxLogs.some(log => log.name === 'Approval' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

    const senderBalanceBefore = await getTokenBalance(tokenContract, sender.address)
    const receiverBalanceBefore = await getTokenBalance(tokenContract, receiver.address)

    const transferTx = await sendShieldedTransaction(
        receiver,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("transferFrom", [sender.address, receiver.address, amountToTransfer])
    )
    const transferTxReceipt = await transferTx.wait()
    const transferTxLogs = transferTxReceipt.logs.map(log => tokenContract.interface.parseLog(log))
    expect(transferTxLogs.some(log => log.name === 'Transfer' && log.args[0] == sender.address && log.args[1] == receiver.address && log.args[2].toNumber() == amountToTransfer)).to.be.true

    const senderBalanceAfter = await getTokenBalance(tokenContract, sender.address)
    const receiverBalanceAfter = await getTokenBalance(tokenContract, receiver.address)

    return [senderBalanceBefore.toNumber(), senderBalanceAfter.toNumber(), receiverBalanceBefore.toNumber(), receiverBalanceAfter.toNumber()]
}

describe('--------Stress Testing----------', () => {
    let tokenContract
    let wallets = [];

    before(async () => {
        console.log("-------Preparing random wallets--------")
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000000)
        await tokenContract.deployed()
        
        const [sender] = await ethers.getSigners()
        const senderBalanceBefore = await getTokenBalance(tokenContract, sender.address)
        console.log("Sender Balance:", senderBalanceBefore)

        // prepare NUM_TESTING_ACCOUNTS accounts and prefund it
        for (let i = 0; i < NUM_TESTING_ACCOUNTS; i++) {
            const wallet = ethers.Wallet.createRandom().connect(ethers.provider)

            const tx = await sender.sendTransaction({
                to: wallet.address,
                value: "1000000000"
            })
            await tx.wait()

            // Transfer ERC20 token
            console.log("Transferring 10000 ERC20 token from", sender.address, "To", wallet.address)
            await transferERC20Token(sender, wallet.address, tokenContract, 10000)
            console.log("Wallet", (i+1), "is ready among", NUM_TESTING_ACCOUNTS, "wallets. Address:", wallet.address)
            const senderBalanceBefore = await getTokenBalance(tokenContract, wallet.address)
            console.log("Balance:", senderBalanceBefore)
            wallets.push(wallet)
        }
    })

    it('Stress ERC20 transfer', async () => {
        const promises = [];
        for (let i = 1 ; i < NUM_TESTING_ACCOUNTS; i++) {
            promises.push(
                transferERC20Token(wallets[i-1], wallets[i].address, tokenContract, 10)
            );
        }

        const res = await Promise.all(promises);
        console.log(res)
    })

    it('Stress ERC20 transferFrom', async () => {
        const promises = [];
        for (let i = 1 ; i < NUM_TESTING_ACCOUNTS; i++) {
            promises.push(transferFromERC20Token(wallets[i-1], wallets[i], tokenContract, 10));
        }
        
        const res = await Promise.all(promises);
        console.log(res)
    })
})