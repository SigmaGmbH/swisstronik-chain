const {expect} = require("chai");

describe('ERC20', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('ERC20Token')
        tokenContract = await ERC20.deploy('test token', 'TT', 10000000000)
        await tokenContract.deployed()
    })

    it('Should return correct name and symbol', async () => {

    })

    it('Should be able to transfer ERC20 tokens', async () => {

    })

    it('Should be able to approve ERC20 tokens', async () => {

    })

    it('Should be able to transferFrom approved tokens', async () => {

    })

    it('Cannot exceed balance during transfer', async () => {

    })

    it('Cannot transfer more than approved', async () => {

    })
})