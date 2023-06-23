const { encryptDataField, decryptNodeResponse } = require('@swisstronik/swisstronik.js')

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