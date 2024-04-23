const { encryptDataField, decryptNodeResponse } = require('@swisstronik/swisstronik.js')

module.exports.sendShieldedTransaction = async (signer, destination, data, value) => {
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
        // gasLimit: 200_000,
        // gasPrice: 0 // We're using 0 gas price in tests. Comment it, if you're running tests on actual network 
    })
}

module.exports.sendShieldedQuery = async (provider, destination, data, value) => {
    // Encrypt call data
    const [encryptedData, usedEncryptedKey] = await encryptDataField(
        provider.connection.url,
        data
    )

    // Do call
    const response = await provider.call({
        to: destination,
        data: encryptedData,
        value
    })

    if (response.startsWith("0x08c379a0")) {
        return response
    }

    // Decrypt call result
    return await decryptNodeResponse(provider.connection.url, response, usedEncryptedKey)
}

// wallet should contain connected provider
module.exports.sendSignedShieldedQuery = async (wallet, destination, data) => {
    if (!wallet.provider) {
        throw new Error("signer doesn't contain connected provider")
    }

    // Encrypt call data
    const [encryptedData, usedEncryptedKey] = await encryptDataField(
        wallet.provider.connection.url,
        data
    )

    // Get chain id for signature
    const networkInfo = await wallet.provider.getNetwork()

    // We treat signed call as a transaction, but it will be sent using eth_call
    const callData = {
        to: destination,
        data: encryptedData,
        chainId: networkInfo.chainId,
    }

    // Extract signature values
    const signedRawCallData = await wallet.signTransaction(callData)
    const decoded = ethers.utils.parseTransaction(signedRawCallData)

    // Construct call with signature values
    const signedCallData = {
        to: decoded.to,
        data: decoded.data,
        v: ethers.utils.hexValue(decoded.v),
        r: ethers.utils.hexValue(decoded.r),
        s: ethers.utils.hexValue(decoded.s),
        chainId: ethers.utils.hexValue(networkInfo.chainId)
    }

    // Do call
    const response = await wallet.provider.send('eth_call', [signedCallData, "latest"])

    // Decrypt call result
    return await decryptNodeResponse(wallet.provider.connection.url, response, usedEncryptedKey)
}