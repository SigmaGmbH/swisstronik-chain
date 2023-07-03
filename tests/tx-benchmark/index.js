require('dotenv').config()
const ethers = require('ethers')
const fs = require('fs')
const { encryptDataField } = require('@swisstronik/swisstronik.js')

const provider = new ethers.providers.JsonRpcProvider(process.env.NODE_RPC || 'http://localhost:8545')
const initialWallet = new ethers.Wallet(process.env.FIRST_PRIVATE_KEY, provider)

const NUM_TESTING_ACCOUNTS = 20

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

    const contract = await factory.deploy('test token', 'TT', 1000000000);
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
        gasPrice: 0 // We're using 0 gas price in tests 
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

    // prepare NUM_TESTING_ACCOUNTS accounts and prefund it
    const wallets = []
    for (let i = 0; i < NUM_TESTING_ACCOUNTS; i++) {
        const wallet = ethers.Wallet.createRandom().connect(provider)

        const tx = await initialWallet.sendTransaction({
            to: wallet.address,
            value: "1000000000"
        })
        await tx.wait()

        // Transfer ERC20 token
        await transferERC20Token(initialWallet, wallet.address, tokenContract, 10000)
        console.log("Wallet", (i + 1), "is ready among", NUM_TESTING_ACCOUNTS, "wallets. Address:", wallet.address)
        wallets.push(wallet)
    }

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

