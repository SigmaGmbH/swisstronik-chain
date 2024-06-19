require('dotenv').config()
const ethers = require('ethers')
const fs = require('fs')
const { encryptDataField } = require('@swisstronik/swisstronik.js')

const provider = new ethers.providers.JsonRpcProvider(process.env.NODE_RPC || 'http://localhost:8545')
const initialWallet = new ethers.Wallet(process.env.FIRST_PRIVATE_KEY, provider)

const INIT_WALLET_SWTR_BALANCE = ethers.utils.parseEther("0.01")
const INIT_WALLET_TOKEN_BALANCE = 1000000000
const NUM_TESTING_ACCOUNTS = 50

async function transferERC20Token(sender, receiverAddress, tokenContract, amountToTransfer) {
    try {
        const tx = await sendShieldedTransaction(
            sender,
            tokenContract.address,
            tokenContract.interface.encodeFunctionData("transfer", [receiverAddress, amountToTransfer])
        )
        await tx.wait()
        return true
    } catch {
        return false
    }
}

// Deploys sample ERC20 contract
async function deployERC20() {
    const metadata = JSON.parse(fs.readFileSync('contracts/ERC20Token.json'))
    const factory = new ethers.ContractFactory(metadata.abi, metadata.bytecode, initialWallet);

    const transferAmount = INIT_WALLET_SWTR_BALANCE.mul(NUM_TESTING_ACCOUNTS);
    const contract = await factory.deploy({value: transferAmount});
    await contract.deployed()

    return contract
}

async function sendShieldedTransaction(signer, destination, data, value) {
    // Encrypt transaction data
    const [encryptedData] = await encryptDataField(
        signer.provider.connection.url,
        data
    )

    // Construct and sign transaction with encrypted data
    return await signer.sendTransaction({
        from: signer.address,
        to: destination,
        data: encryptedData,
        value,
        // gasLimit: 300_000,
        // gasPrice: 7 // We're using 0 gas price in tests 
    })
}

function parseHrtimeToSeconds(hrtime) {
    return (hrtime[0] + (hrtime[1] / 1e9)).toFixed(5);
}

function outputResult(res, elapsedTime) {
    const succeed = res.filter(e => e)
    const successRate = Math.floor(succeed.length / res.length * 100)
    console.log(`Success rate: ${successRate}%`)
    console.log(`Time per tx: `, elapsedTime / res.length)
}


async function main() {
    console.log('Deploying test ERC20')
    const tokenContract = await deployERC20()
    console.log(`ERC20 deployed with address: ${tokenContract.address}`)

    console.log('Initializing wallets')
    const wallets = []
    for (let i = 0; i < NUM_TESTING_ACCOUNTS; i++) {
        const wallet = ethers.Wallet.createRandom().connect(provider)
        wallets.push(wallet)
    }
    const walletAddresses = wallets.map((wallet) => wallet.address)
    const amounts = [...Array(NUM_TESTING_ACCOUNTS)].map(() => INIT_WALLET_SWTR_BALANCE)
    const swtrTransferTx = await sendShieldedTransaction(
        initialWallet,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("bulkTransfer", [walletAddresses, amounts])
    )
    await swtrTransferTx.wait()

    const tokenAmounts = [...Array(NUM_TESTING_ACCOUNTS)].map(() => INIT_WALLET_TOKEN_BALANCE)
    const tokenTransferTx = await sendShieldedTransaction(
        initialWallet,
        tokenContract.address,
        tokenContract.interface.encodeFunctionData("bulkMint", [walletAddresses, tokenAmounts])
    )
    await tokenTransferTx.wait()

    console.log('Wallets are ready')

    const startTime = process.hrtime();
    const promises = [];
    for (let i = 1 ; i < NUM_TESTING_ACCOUNTS; i++) {
        promises.push(
            transferERC20Token(wallets[i-1], wallets[i].address, tokenContract, 10)
        );
    }

    const res = await Promise.all(promises);
    const elapsedSeconds = parseHrtimeToSeconds(process.hrtime(startTime));
    outputResult(res, elapsedSeconds)
}

main()

