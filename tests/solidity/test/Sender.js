const { expect } = require('chai')
const { sendSignedShieldedQuery, sendShieldedQuery } = require("./testUtils")
const { ethers } = require("hardhat")

describe('Recover sender in query', () => {
    let signer, senderContract

    before(async () => {
        const [ethersSigner] = await ethers.getSigners()
        signer = ethersSigner

        const SenderContract = await ethers.getContractFactory('Sender')
        senderContract = await SenderContract.deploy()
        await senderContract.deployed()
    })

    it('Should return empty msg.sender if was not signed', async () => {
        const emptySenderRequest = await sendShieldedQuery(
            ethers.provider,
            senderContract.address,
            senderContract.interface.encodeFunctionData("getSender", [])
        );
    
        const emptySenderResult = senderContract.interface.decodeFunctionResult("getSender", emptySenderRequest)[0]
        expect(emptySenderResult).to.be.equal(ethers.constants.AddressZero)
    })

    it('Should return signer address', async () => {
        const senderWallet = ethers.Wallet.createRandom().connect(signer.provider)
        const req = await sendSignedShieldedQuery(
            senderWallet,
            senderContract.address,
            senderContract.interface.encodeFunctionData("getSender", []),
            0
        );
    
        const result = senderContract.interface.decodeFunctionResult("getSender", req)[0]
        expect(result).to.be.equal(senderWallet.address)
    })
})
