const hre = require("hardhat");
const { encryptDataField, decryptNodeResponse } = require('@swisstronik/swisstronik.js')

const getTokenBalance = async (signer, contract) => {
    const messageHash = ethers.utils.solidityKeccak256(["address"], [signer.address])
    const messageHashBinary = ethers.utils.arrayify(messageHash)
    const signature = await signer.signMessage(messageHashBinary)

    const balanceResponse = await sendShieldedQuery(
        signer.provider,
        contract.address,
        contract.interface.encodeFunctionData("balanceOfWithSignature", [signer.address, signature])
    );
    return contract.interface.decodeFunctionResult("balanceOfWithSignature", balanceResponse)[0]
}

const sendShieldedQuery = async (provider, destination, data) => {
    // Encrypt call data
    const [encryptedData, usedEncryptedKey] = await encryptDataField(
        provider.connection.url,
        data
    )

    // Do call
    const response = await provider.call({
        to: destination,
        data: encryptedData,
    })

    // Decrypt call result
    return await decryptNodeResponse(provider.connection.url, response, usedEncryptedKey)
}

const sendShieldedTransaction = async (signer, destination, data, value) => {
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
    })
}

async function main() {
    const amountToSend = 10000
    const [sender] = await hre.ethers.getSigners()
    const contractAddress = process.env.ADDRESS
    const receiver = await hre.ethers.Wallet.createRandom().connect(hre.ethers.provider)

    console.log(`Send ${amountToSend} uswtr to ${contractAddress}`)

    // Send 100 uswtr to convert them to PSWTR
    const tx = await sender.sendTransaction({
        from: sender.address,
        to: contractAddress,
        value: amountToSend
    })
    await tx.wait()

    // Construct contract and check balance
    const factory = await hre.ethers.getContractFactory('PrivateSWTR')
    const contract = factory.attach(contractAddress)
    const tokenBalance = await getTokenBalance(sender, contract)
    console.log(`Now you have ${tokenBalance} PSWTR`)

    // Send PSWTR to random wallet
    const transferTx = await sendShieldedTransaction(
        sender,
        contract.address,
        contract.interface.encodeFunctionData("transfer", [receiver.address, amountToSend])
    )
    await transferTx.wait()

    // Check receiver balance
    const receiverBalance = await getTokenBalance(receiver, contract)
    console.log(`Receiver has ${receiverBalance} PSWTR`)
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});