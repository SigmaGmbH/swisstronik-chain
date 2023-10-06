const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Verifiable credential', () => {
    let vc

    before(async () => {
        const VC = await ethers.getContractFactory('VC')
        vc = await VC.deploy()
        await vc.deployed()
    })

    it('Should verify VC', async () => {
        const [signer] = await ethers.getSigners()

        const isAuthorizedResponse = await sendShieldedQuery(
            signer.provider,
            vc.address,
            vc.interface.encodeFunctionData("isAuthorized", [])
        );
        const isAuthorizedBefore = counterContract.interface.decodeFunctionResult("isAuthorized", isAuthorizedResponse)
    })
})