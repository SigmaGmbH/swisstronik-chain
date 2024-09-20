const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("BlockInfo", function () {
    describe('Unencrypted',() => {
        const unencryptedProvider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
        const unencryptedSigner = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", unencryptedProvider)

        let blockInfoContract

        before(async () => {
            const factory = await ethers.getContractFactory('BlockInfo')
            blockInfoContract = await factory.connect(unencryptedSigner).deploy()
            await blockInfoContract.connect(unencryptedProvider).deployed()
        })

        it("Should retrieve block base fee", async () => {
            const baseFee = await blockInfoContract.connect(unencryptedProvider).getBlockBaseFee();
            expect(baseFee).to.be.gt(0);
        });

        it("Should retrieve block chain id", async () => {
            const chainId = await blockInfoContract.connect(unencryptedProvider).getBlockChainId();
            const providerChainId = await unencryptedProvider.getNetwork().then(n => n.chainId)
            expect(chainId).to.equal(providerChainId);
        });

        it("Should retrieve block coinbase", async () => {
            const coinbase = await blockInfoContract.connect(unencryptedProvider).getBlockCoinbase();
            expect(ethers.utils.isAddress(coinbase)).to.be.true;
        });

        it("Should retrieve block difficulty", async () => {
            const difficulty = await blockInfoContract.connect(unencryptedProvider).getBlockDifficulty();
            expect(difficulty).to.be.equal(0);
        });

        it("Should retrieve block gas limit", async () => {
            const gasLimit = await blockInfoContract.connect(unencryptedProvider).getBlockGasLimit();
            expect(gasLimit).to.be.gt(0);
        });

        it("Should retrieve block number", async () => {
            const blockNumber = await blockInfoContract.connect(unencryptedProvider).getBlockNumber();
            expect(blockNumber).to.be.gt(0);
        });

        it("Should retrieve block timestamp", async () => {
            const timestamp = await blockInfoContract.connect(unencryptedProvider).getBlockTimestamp();
            expect(timestamp).to.be.gt(0);
        });

        it("Should retrieve blockhash of a previous block", async () => {
            const currentBlock = await unencryptedProvider.getBlockNumber();
            const blockHash = await blockInfoContract.connect(unencryptedProvider).getBlockhash(currentBlock - 1);
            expect(blockHash).to.not.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
        });

        describe("Edge cases", function () {
            it("Should return zero for block hash of a future block", async function () {
                const currentBlock = await unencryptedProvider.getBlockNumber();
                const futureBlockHash = await blockInfoContract.connect(unencryptedProvider).getBlockhash(currentBlock + 1000);
                expect(futureBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
            });

            it("Should return zero for block hash of a block more than 256 blocks ago", async function () {
                const currentBlock = await unencryptedProvider.getBlockNumber();
                if (currentBlock > 256) {
                    const oldBlockHash = await blockInfoContract.connect(unencryptedProvider).getBlockhash(currentBlock - 257);
                    expect(oldBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
                }
            });
        });
    })
});