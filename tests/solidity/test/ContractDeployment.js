const { expect } = require("chai");
const { ethers } = require("hardhat")

describe('Contract Deployment', () => {
    it("Should deploy to an address derived from deployer's nonce", async function () {
        const [deployer] = await ethers.getSigners();
        const nonce = await deployer.getTransactionCount();
        const expectedAddress = ethers.utils.getContractAddress({
            from: deployer.address,
            nonce: nonce
        });
        const expectedSignerNonce = nonce + 1
        const expectedContractNonce = 1

        const ContractFactory = await ethers.getContractFactory("Counter");
        const contract = await ContractFactory.deploy();
        await contract.deployed();

        const signerNonceAfter = await deployer.getTransactionCount()
        const contractNonce = await deployer.provider.getTransactionCount(contract.address)

        expect(contract.address.toLowerCase()).to.equal(
            expectedAddress.toLowerCase()
        );
        expect(signerNonceAfter).to.equal(expectedSignerNonce)
        expect(contractNonce).to.equal(expectedContractNonce)
    });
})