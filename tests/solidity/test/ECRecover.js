const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("ECRecover Test", function () {
    let contract;
    let signer;

    before(async function () {
        const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547');
        // WARNING: DO NOT USE THIS PRIVATE KEY, SINCE IT IS PUBLICLY AVAILABLE
        signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)
        const ECRecoverTest = await ethers.getContractFactory("ECRecoverPrecompileTest", signer);
        contract = await ECRecoverTest.deploy();
        await contract.deployed();
    });

    it("should recover the correct address from a valid signature", async function () {
        const message = "Hello, Ethereum!";
        const messageHash = ethers.utils.hashMessage(message);
        const signature = await signer.signMessage(message);
        const r = signature.slice(0, 66);
        const s = "0x" + signature.slice(66, 130);
        const v = parseInt(signature.slice(130, 132), 16);
        const recoveredAddress = await contract.recoverAddress(messageHash, v, r, s);
        expect(recoveredAddress).to.equal(signer.address);
    });

    it("should recover different address from an invalid signature", async function () {
        const message = "Hello, Ethereum!";
        const invalidMessageHash = ethers.utils.hashMessage("Invalid message");
        const signature = await signer.signMessage(message);

        const r = signature.slice(0, 66);
        const s = "0x" + signature.slice(66, 130);
        const v = parseInt(signature.slice(130, 132), 16);

        const recoveredAddress = await contract.recoverAddress(invalidMessageHash, v, r, s)
        // Expected value is obtained from geth node
        expect(recoveredAddress).to.be.equal("0x3e52eB4a136Df5D507B6e47Fc784424eeF1E94fC");
    });

    it("should accept high s-values", async function () {
        const messageHash = "0x18c547e4f7b0f325ad1e56f57e26c745b09a3e503d86e00e5255ff7f715d3d1c";
        const v = 28;
        const r = "0x73b1693892219d736caba55bdb67216e485557ea6b6af75f37096c9aa6a5a75f";
        const s = "0xeeb940b1d03b21e36b0e47e79769f095fe2ab855bd91e3a38756b7d75a9c4549";

        const recoveredAddress = await contract.recoverAddress(messageHash, v, r, s)
        // Expected value is obtained from geth node
        expect(recoveredAddress).to.be.equal("0xa94f5374Fce5edBC8E2a8697C15331677e6EbF0B");
    });
});