const {expect} = require("chai");

describe('ERC721', () => {
    let nftContract

    before(async () => {
        const ERC721 = await ethers.getContractFactory('ERC721Token')
        nftContract = await ERC721.deploy('test token', 'TT')
        await nftContract.deployed()
    })

    it('Should return correct token name and symbol', async () => {

    })

    it('Should be able to mint new NFT', async () => {

    })

    it('Should be able to transfer an NFT', async () => {

    })

    it('Cannot transfer unapproved NFT', async () => {

    })

    it('Should return metadata URI for NFT', async () => {

    })
})