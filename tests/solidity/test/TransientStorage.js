const { expect } = require("chai");
const { ethers } = require("hardhat");

const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)
const otherSigner = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", provider)

describe("Transient storage", function () {
    let mulService, mulCaller;

    before(async () => {
        // Deploy MulService
        const MulService = await ethers.getContractFactory("MulService", signer);
        mulService = await MulService.deploy();
        await mulService.deployed();

        // Deploy MulCaller
        const MulCaller = await ethers.getContractFactory("MulCaller", signer);
        mulCaller = await MulCaller.deploy(mulService.address);
        await mulCaller.deployed();
    });

    it("should correctly set multiplier and multiply", async () => {
        const multiplier = 5;
        const value = 10;
        const expectedResult = multiplier * value;

        const result = await mulCaller.callStatic.runMultiply(multiplier, value);

        expect(result).to.equal(expectedResult);
    });

    it("should use transient storage (multiplier not persisting between calls)", async () => {
        const multiplier1 = 5;
        const value1 = 10;
        const multiplier2 = 3;
        const value2 = 7;

        const tx = await mulCaller.runMultiply(multiplier1, value1);
        await tx.wait()

        const result = await mulCaller.callStatic.runMultiply(multiplier2, value2);

        expect(result).to.equal(multiplier2 * value2);
    });

    it("should handle zero multiplier", async () => {
        const multiplier = 0;
        const value = 100;

        const result = await mulCaller.callStatic.runMultiply(multiplier, value);

        expect(result).to.equal(0);
    });

});