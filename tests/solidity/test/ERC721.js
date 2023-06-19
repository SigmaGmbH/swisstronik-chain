require('dotenv').config()
const { expect } = require("chai")
const { sendShieldedTransaction, sendShieldedQuery, getProvider } = require("./testUtils")

const getTokenBalance = async (provider, privateKey, contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        provider,
        privateKey,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

const getOwnerOf = async (provider, privateKey, contract, itemId) => {
    const ownerResponse = await sendShieldedQuery(
        provider,
        privateKey,
        contract.address,
        contract.interface.encodeFunctionData("ownerOf", [itemId])
    );
    return contract.interface.decodeFunctionResult("ownerOf", ownerResponse)[0]
}

describe('ERC721', () => {
    let nftContract
    const provider = getProvider()
    const senderPrivateKey = process.env.FIRST_PRIVATE_KEY
    const receiverPrivateKey = process.env.SECOND_PRIVATE_KEY

    before(async () => {
        const ERC721 = await ethers.getContractFactory('ERC721Token')
        nftContract = await ERC721.deploy('test token', 'TT', {gasLimit: 2500000})
        await nftContract.deployed()
    })

    it('Should return correct token name and symbol', async () => {
        const nameResponse = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            nftContract.address,
            nftContract.interface.encodeFunctionData("name", [])
        );
        const name = nftContract.interface.decodeFunctionResult("name", nameResponse)[0]
        expect(name).to.be.equal('test token')

        const symbolResponse = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            nftContract.address,
            nftContract.interface.encodeFunctionData("symbol", [])
        );
        const symbol = nftContract.interface.decodeFunctionResult("symbol", symbolResponse)[0]
        expect(symbol).to.be.equal('TT')
    })

    it('Should be able to mint new NFT', async () => {
        const [user] = await ethers.getSigners()
        const tokenURI = "http://nftstorage.com/item/1"
        const expectedItemId = 0

        const tx = await sendShieldedTransaction(
            provider,
            senderPrivateKey,
            nftContract.address,
            nftContract.interface.encodeFunctionData("createItem", [user.address, tokenURI])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => nftContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Transfer' && log.args[0] == ethers.constants.AddressZero && log.args[1] == user.address && log.args[2].toNumber() == expectedItemId)).to.be.true

        const userBalance = await getTokenBalance(provider, senderPrivateKey, nftContract, user.address)
        expect(userBalance).to.be.equal(1)

        const ownerOfItem = await getOwnerOf(provider, senderPrivateKey, nftContract, expectedItemId)
        expect(ownerOfItem).to.be.equal(user.address)
    })

    it('Should be able to transfer an NFT', async () => {
        const [user, receiver] = await ethers.getSigners()
        const itemId = 0

        const tx = await sendShieldedTransaction(
            provider,
            senderPrivateKey,
            nftContract.address,
            nftContract.interface.encodeFunctionData("transferFrom", [user.address, receiver.address, itemId])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => nftContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Transfer' && log.args[0] == user.address && log.args[1] == receiver.address && log.args[2].toNumber() == itemId)).to.be.true

        const userBalance = await getTokenBalance(provider, senderPrivateKey, nftContract, user.address)
        expect(userBalance).to.be.equal(0)

        const receiverBalance = await getTokenBalance(provider, receiverPrivateKey, nftContract, receiver.address)
        expect(receiverBalance).to.be.equal(1)

        const ownerOfItem = await getOwnerOf(provider, receiverPrivateKey, nftContract, itemId)
        expect(ownerOfItem).to.be.equal(receiver.address)
    })

    it('Cannot transfer unapproved NFT', async () => {
        const itemId = 0
        const [wrongSender, receiver] = await ethers.getSigners()

        let failed = false
        try {
            await sendShieldedTransaction(
                provider,
                senderPrivateKey,
                tokenContract.address,
                tokenContract.interface.encodeFunctionData("transferFrom", [receiver.address, wrongSender.address, itemId])
            )
        } catch {
            failed = true
        }

        expect(failed).to.be.true
    })

    it('Should return metadata URI for NFT', async () => {
        const itemId = 0
        const metadataURIResponse = await sendShieldedQuery(
            provider,
            senderPrivateKey,
            nftContract.address,
            nftContract.interface.encodeFunctionData("tokenURI", [itemId])
        );
        const metadataURI = nftContract.interface.decodeFunctionResult("tokenURI", metadataURIResponse)[0]
        expect(metadataURI).to.be.equal("http://nftstorage.com/item/1")
    })
})