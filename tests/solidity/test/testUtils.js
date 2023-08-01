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
        gasPrice: 0 // We're using 0 gas price in tests. Comment it, if you're running tests on actual network 
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

    // Decrypt call result
    return await decryptNodeResponse(provider.connection.url, response, usedEncryptedKey)
}