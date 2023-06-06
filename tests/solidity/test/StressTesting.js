require('dotenv').config()
const { expect } = require("chai")
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery, getProvider } = require("./testUtils")

const NUM_TESTING_ACCOUNTS = 50;

const getTokenBalance = async (provider, privateKey, contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        provider,
        privateKey,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

const transferERC20Token = async (provider, sender, receiver, tokenContract, amountToTransfer) => {
    const senderPrivateKey = sender.privateKey
    const receiverPrivateKey = receiver.privateKey

    const senderBalanceBefore = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
    const receiverBalanceBefore = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

    const tx = await sendShieldedTransaction(
        provider,
        senderPrivateKey,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("transfer", [receiver.address, amountToTransfer])
    )
    await tx.wait()
    const senderBalanceAfter = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
    const receiverBalanceAfter = await getTokenBalance(provider, receiverPrivateKey, tokenContract, receiver.address)

    return [senderBalanceBefore.toNumber(), senderBalanceAfter.toNumber(), receiverBalanceBefore.toNumber(), receiverBalanceAfter.toNumber()]
}

const transferFromERC20Token  = async (provider, sender, receiver, tokenContract, amountToTransfer) =>{
    const senderPrivateKey = sender.privateKey
    const receiverPrivateKey = receiver.privateKey

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

    return [senderBalanceBefore.toNumber(), senderBalanceAfter.toNumber(), receiverBalanceBefore.toNumber(), receiverBalanceAfter.toNumber()]
}

describe('--------Stress Testing----------', () => {
    let tokenContract
    const provider = getProvider()
    const senderPrivateKey = process.env.FIRST_PRIVATE_KEY

    let wallets = [];

    before(async () => {
        console.log("-------Preparing random wallets--------")
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000000)
        await tokenContract.deployed()
        
        const [sender] = await ethers.getSigners()
        const senderBalanceBefore = await getTokenBalance(provider, senderPrivateKey, tokenContract, sender.address)
        console.log("Sender Balance:", senderBalanceBefore)

        // prepare NUM_TESTING_ACCOUNTS accounts and prefund it
        for (let i = 0; i < NUM_TESTING_ACCOUNTS; i++) {
            const wallet = ethers.Wallet.createRandom()
            const tx = await sender.sendTransaction({
                to: wallet.address,
                value: "1000000000",
                gasLimit: 21000,
                gasPrice: 10,
            })
            await tx.wait()

            // Transfer ERC20 token
            console.log("Transferring 10000 ERC20 token from", sender.address, "To", wallet.address)
            await transferERC20Token(provider, {...sender, privateKey: senderPrivateKey}, wallet, tokenContract, 10000)
            console.log("Wallet", (i+1), "is ready among", NUM_TESTING_ACCOUNTS, "wallets. Address:", wallet.address)
            const senderBalanceBefore = await getTokenBalance(provider, wallet.privateKey, tokenContract, wallet.address)
            console.log("Balance:", senderBalanceBefore)
            wallets.push(wallet)
        }
    })

    it('Stress ERC20 transfer', async () => {
        const promises = [];
        for (let i = 1 ; i < NUM_TESTING_ACCOUNTS; i++) {
            promises.push(
                transferERC20Token(provider, wallets[i-1], wallets[i], tokenContract, 10)
            );
        }

        const res = await Promise.all(promises);
        console.log(res)
    })

    it('Stress ERC20 transferFrom', async () => {
        const promises = [];
        for (let i = 1 ; i < NUM_TESTING_ACCOUNTS; i++) {
            promises.push(transferFromERC20Token(provider, wallets[i-1], wallets[i], tokenContract, 10));
        }
        
        const res = await Promise.all(promises);
        console.log(res)
    })
})