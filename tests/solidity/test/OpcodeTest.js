const { expect } = require('chai')
const {ethers} = require("hardhat");

const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)

describe('OPCODE test', () => {
    let opcodeTest

    before(async () => {
        const factory = await ethers.getContractFactory('OpcodeTest', signer)
        opcodeTest = await factory.deploy()
        await opcodeTest.deployed()
    })

    it('Should throw invalid op code', async () => {
        await expect(opcodeTest.testInvalid()).to.be.rejectedWith("InvalidOpcode(Opcode(254))")
    })

    it('Should revert', async () => {
        await expect(opcodeTest.testRevert()).to.be.rejectedWith("reverted")
    })

    it("should correctly perform addition", async function () {
        expect(await opcodeTest.testAdd(5, 3)).to.equal(8);
    });

    it("should correctly perform subtraction", async function () {
        expect(await opcodeTest.testSub(10, 4)).to.equal(6);
    });

    it("should correctly perform multiplication", async function () {
        expect(await opcodeTest.testMul(7, 6)).to.equal(42);
    });

    it("should correctly perform division", async function () {
        expect(await opcodeTest.testDiv(20, 5)).to.equal(4);
    });

    it("should correctly perform modulo", async function () {
        expect(await opcodeTest.testMod(17, 5)).to.equal(2);
    });

    it("should correctly perform left shift", async function () {
        expect(await opcodeTest.testShl(1, 2)).to.equal(4);
    });

    it("should correctly perform right shift", async function () {
        expect(await opcodeTest.testShr(8, 2)).to.equal(2);
    });

    it("should correctly perform bitwise AND", async function () {
        expect(await opcodeTest.testAnd(12, 5)).to.equal(4);
    });

    it("should correctly perform bitwise OR", async function () {
        expect(await opcodeTest.testOr(12, 5)).to.equal(13);
    });

    it("should correctly perform bitwise XOR", async function () {
        expect(await opcodeTest.testXor(12, 5)).to.equal(9);
    });

    it("should correctly perform bitwise NOT", async function () {
        expect(await opcodeTest.testNot(0)).to.equal(ethers.constants.MaxUint256);
    });

    it("should correctly perform SSTORE operation", async function () {
        const tx = await opcodeTest.testSSTORE(42);
        await tx.wait()

        expect(await opcodeTest.storedValue()).to.equal(42);
    });

    it("should correctly perform MSTORE operation", async function () {
        expect(await opcodeTest.testMSTORE()).to.equal(42);
    });

    it("should correctly perform EXTCODESIZE operation", async function () {
        const size = await opcodeTest.testEXTCODESIZE(opcodeTest.address);
        expect(size).to.be.gt(0);
    });

    it("should return 0 for EXTCODESIZE of EOA", async function () {
        const size = await opcodeTest.testEXTCODESIZE(signer.address);
        expect(size).to.equal(0);
    });
});