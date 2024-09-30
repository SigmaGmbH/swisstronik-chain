const { expect } = require("chai");
const { ethers } = require("hardhat");
const { sendShieldedQuery } = require("./testUtils")

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

        // it("Should retrieve block gas limit", async () => {
        //     const gasLimit = await blockInfoContract.connect(unencryptedProvider).getBlockGasLimit();
        //     expect(gasLimit).to.be.gt(0);
        // });

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
            it("Should return zero for block hash of a future block", async () => {
                const currentBlock = await unencryptedProvider.getBlockNumber();
                const futureBlockHash = await blockInfoContract.connect(unencryptedProvider).getBlockhash(currentBlock + 1000);
                expect(futureBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
            });

            it("Should return zero for block hash of a block more than 256 blocks ago", async () => {
                const currentBlock = await unencryptedProvider.getBlockNumber();
                if (currentBlock > 256) {
                    const oldBlockHash = await blockInfoContract.connect(unencryptedProvider).getBlockhash(currentBlock - 257);
                    expect(oldBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
                }
            });
        });
    })

    describe('Encrypted',() => {
        let blockInfoContract, signer

        before(async () => {
            const factory = await ethers.getContractFactory('BlockInfo')
            blockInfoContract = await factory.deploy()
            await blockInfoContract.deployed()

            const [ethersSigner] = await ethers.getSigners()
            signer = ethersSigner
        })

        it("Should retrieve block base fee", async () => {
            const baseFeeRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockBaseFee", [])
            );
            const baseFee = blockInfoContract.interface.decodeFunctionResult("getBlockBaseFee", baseFeeRes)[0]

            expect(baseFee).to.be.gt(0)
        });

        it("Should retrieve block chain id", async () => {
            const [signer] = await ethers.getSigners()
            const providerChainId = await signer.provider.getNetwork().then(n => n.chainId)

            const chainIdRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockChainId", [])
            );
            const chainId = blockInfoContract.interface.decodeFunctionResult("getBlockChainId", chainIdRes)[0]

            expect(chainId).to.equal(providerChainId)
        });

        it("Should retrieve block coinbase", async () => {
            const coinbaseRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockCoinbase", [])
            );
            const coinbase = blockInfoContract.interface.decodeFunctionResult("getBlockCoinbase", coinbaseRes)[0]

            expect(ethers.utils.isAddress(coinbase)).to.be.true
        });

        it("Should retrieve block difficulty", async () => {
            const difficultyRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockDifficulty", [])
            );
            const difficulty = blockInfoContract.interface.decodeFunctionResult("getBlockDifficulty", difficultyRes)[0]

            expect(difficulty).to.be.equal(0)
        });

        // it("Should retrieve block gas limit", async () => {
        //     const gasLimitRes = await sendShieldedQuery(
        //         signer.provider,
        //         blockInfoContract.address,
        //         blockInfoContract.interface.encodeFunctionData("getBlockGasLimit", [])
        //     );
        //     const gasLimit = blockInfoContract.interface.decodeFunctionResult("getBlockGasLimit", gasLimitRes)[0]
        //
        //     expect(gasLimit).to.be.gt(0)
        // });

        it("Should retrieve block number", async () => {
            const blockNumberRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockNumber", [])
            );
            const blockNumber = blockInfoContract.interface.decodeFunctionResult("getBlockNumber", blockNumberRes)[0]

            expect(blockNumber).to.be.gt(0)
        });

        it("Should retrieve block timestamp", async () => {
            const timestampRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockTimestamp", [])
            );
            const timestamp = blockInfoContract.interface.decodeFunctionResult("getBlockTimestamp", timestampRes)[0]

            expect(timestamp).to.be.gt(0)
        });

        it("Should retrieve blockhash of a previous block", async () => {
            const currentBlock = await signer.provider.getBlockNumber();
            const blockHashRes = await sendShieldedQuery(
                signer.provider,
                blockInfoContract.address,
                blockInfoContract.interface.encodeFunctionData("getBlockhash", [currentBlock - 1])
            );
            const blockHash = blockInfoContract.interface.decodeFunctionResult("getBlockhash", blockHashRes)[0]

            expect(blockHash).to.not.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
        });

        describe("Edge cases", function () {
            it("Should return zero for block hash of a future block", async () => {
                const currentBlock = await signer.provider.getBlockNumber();

                const blockHashRes = await sendShieldedQuery(
                    signer.provider,
                    blockInfoContract.address,
                    blockInfoContract.interface.encodeFunctionData("getBlockhash", [currentBlock + 1000])
                );
                const futureBlockHash = blockInfoContract.interface.decodeFunctionResult("getBlockhash", blockHashRes)[0]

                expect(futureBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
            });

            it("Should return zero for block hash of a block more than 256 blocks ago", async () => {
                const currentBlock = await signer.provider.getBlockNumber();
                if (currentBlock > 256) {
                    const blockHashRes = await sendShieldedQuery(
                        signer.provider,
                        blockInfoContract.address,
                        blockInfoContract.interface.encodeFunctionData("getBlockhash", [currentBlock - 257])
                    );
                    const oldBlockHash = blockInfoContract.interface.decodeFunctionResult("getBlockhash", blockHashRes)[0]
                    expect(oldBlockHash).to.equal('0x0000000000000000000000000000000000000000000000000000000000000000');
                }
            });
        });
    })
});