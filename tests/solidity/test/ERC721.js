const {expect} = require("chai");
const {ethers} = require("hardhat")

describe('ERC721', () => {
    let nftContract

    before(async () => {
        const ERC721 = await ethers.getContractFactory('ERC721Token')
        nftContract = await ERC721.deploy('test token', 'TT')
        await nftContract.deployed()
    })

    it('Should return correct token name and symbol', async () => {
        const name = await nftContract.name()
        const symbol = await nftContract.symbol()

        expect(name).to.be.equal('test token')
        expect(symbol).to.be.equal('TT')
    })

    it('Should be able to mint new NFT', async () => {
        const [user] = await ethers.getSigners()
        const tokenURI = "http://nftstorage.com/item/1"
        const expectedItemId = 0

        await expect(nftContract.connect(user).createItem(user.address, tokenURI))
            .to.emit(nftContract, 'Transfer')
            .withArgs(ethers.constants.AddressZero, user.address, expectedItemId)

        const userBalance = await nftContract.balanceOf(user.address)
        expect(userBalance).to.be.equal(1)

        const ownerOfItem = await nftContract.ownerOf(expectedItemId)
        expect(ownerOfItem).to.be.equal(user.address)
    })

    it('Should be able to transfer an NFT', async () => {
        const [user, receiver] = await ethers.getSigners()
        const itemId = 0

        await expect(nftContract.connect(user).transferFrom(user.address, receiver.address, itemId))
            .to.emit(nftContract, 'Transfer')
            .withArgs(user.address, receiver.address, itemId)

        const userBalance = await nftContract.balanceOf(user.address)
        expect(userBalance).to.be.equal(0)

        const receiverBalance = await nftContract.balanceOf(receiver.address)
        expect(receiverBalance).to.be.equal(1)

        const ownerOfItem = await nftContract.ownerOf(itemId)
        expect(ownerOfItem).to.be.equal(receiver.address)
    })

    it('Cannot transfer unapproved NFT', async () => {
        const itemId = 0
        const [wrongSender, receiver] = await ethers.getSigners()
        const tx = await nftContract.connect(wrongSender).transferFrom(receiver.address, wrongSender.address, itemId)
        await expect(tx.wait()).to.be.rejected
    })

    it('Should return metadata URI for NFT', async () => {
        const itemId = 0
        const metadataURI = await nftContract.tokenURI(itemId)
        expect(metadataURI).to.be.equal("http://nftstorage.com/item/1")
    })
})