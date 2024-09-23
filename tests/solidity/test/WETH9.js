const { expect } = require("chai");
const { ethers } = require("hardhat");

const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)
const receiver = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", provider)

describe("WETH9", function() {
    let weth;

    before(async function() {
        const factory = await ethers.getContractFactory("WETH9", signer);
        weth = await factory.deploy();
        await weth.deployed();
    });

    describe("Deployment", function() {
        it("should set the correct name, symbol, and decimals", async function() {
            expect(await weth.name()).to.equal("Wrapped Ether");
            expect(await weth.symbol()).to.equal("WETH");
            expect(await weth.decimals()).to.equal(18);
        });
    });

    describe("Deposit", function() {
        it("should accept deposits and update balance", async function() {
            const depositAmount = ethers.utils.parseEther("1");
            await expect(weth.connect(signer).deposit({ value: depositAmount }))
                .to.emit(weth, "Deposit")
                .withArgs(signer.address, depositAmount);

            const balance = await weth.balanceOf(signer.address);
            expect(balance).to.equal(depositAmount);
        });

        it("should accept deposits via fallback function", async function() {
            const balanceBefore = await weth.balanceOf(signer.address)

            const depositAmount = ethers.utils.parseEther("1");
            const tx = await signer.sendTransaction({
                to: weth.address,
                value: depositAmount
            });
            await tx.wait()

            const balanceAfter = await weth.balanceOf(signer.address)
            expect(balanceAfter).to.equal(balanceBefore.add(depositAmount));
        });
    });

    describe("Withdraw", function() {
        it("should allow withdrawals and update balance", async function() {
            const balanceBefore = await weth.balanceOf(signer.address)

            const tx = await weth.connect(signer).withdraw(balanceBefore)
            const receipt = await tx.wait()

            const logs = receipt.logs.map(log => weth.interface.parseLog(log))
            expect(logs.length).to.be.equal(1)
            expect(logs[0].name).to.be.equal('Withdrawal')
            expect(logs[0].args[0]).to.be.equal(signer.address)
            expect(logs[0].args[1]).to.be.equal(balanceBefore)

            const balanceAfter = await weth.balanceOf(signer.address);
            expect(balanceAfter).to.equal(0);
        });

        it("should revert when trying to withdraw more than balance", async function() {
            const withdrawAmount = ethers.utils.parseEther("10000000");
            await expect(weth.connect(signer).withdraw(withdrawAmount))
                .to.be.rejectedWith("reverted");
        });
    });

    describe("Total Supply", function() {
        it("should return the correct total supply", async function() {
            const totalSupplyBefore = await weth.totalSupply()
            const depositAmount = ethers.utils.parseEther("1")

            const tx = await weth.connect(signer).deposit({ value: depositAmount });
            await tx.wait()

            const totalSupplyAfter = await weth.totalSupply();
            expect(totalSupplyAfter).to.equal(totalSupplyBefore.add(depositAmount));
        });
    });

    describe("Approve and TransferFrom", function() {
        it("should approve and allow transferFrom", async function() {
            const depositAmount = ethers.utils.parseEther("1");
            const transferAmount = ethers.utils.parseEther("0.5");

            const depositTx = await weth.connect(signer).deposit({ value: depositAmount });
            await depositTx.wait()

            const senderBalanceBefore = await weth.balanceOf(signer.address)
            const receiverBalanceBefore = await weth.balanceOf(receiver.address)

            const approvalTx = await weth.connect(signer).approve(receiver.address, transferAmount)
            const approvalReceipt = await approvalTx.wait()

            const approvalLogs = approvalReceipt.logs.map(log => weth.interface.parseLog(log))
            expect(approvalLogs.length).to.be.equal(1)
            expect(approvalLogs[0].name).to.be.equal('Approval')
            expect(approvalLogs[0].args[0]).to.be.equal(signer.address)
            expect(approvalLogs[0].args[1]).to.be.equal(receiver.address)
            expect(approvalLogs[0].args[2]).to.be.equal(transferAmount)

            const transferFromTx = await weth.connect(receiver).transferFrom(signer.address, receiver.address, transferAmount)
            const transferFromReceipt = await transferFromTx.wait()
            const transferFromLogs = transferFromReceipt.logs.map(log => weth.interface.parseLog(log))
            expect(transferFromLogs.length).to.be.equal(1)
            expect(transferFromLogs[0].name).to.be.equal('Transfer')
            expect(transferFromLogs[0].args[0]).to.be.equal(signer.address)
            expect(transferFromLogs[0].args[1]).to.be.equal(receiver.address)
            expect(transferFromLogs[0].args[2]).to.be.equal(transferAmount)

            const senderBalanceAfter = await weth.balanceOf(signer.address)
            const receiverBalanceAfter = await weth.balanceOf(receiver.address)

            expect(senderBalanceAfter).to.equal(senderBalanceBefore.sub(transferAmount));
            expect(receiverBalanceAfter).to.equal(receiverBalanceBefore.add(transferAmount));
        });
    });

    describe("Transfer", function() {
        it("should transfer tokens between accounts", async function() {
            const depositAmount = ethers.utils.parseEther("1");
            const tx = await weth.connect(signer).deposit({ value: depositAmount });
            await tx.wait()

            const senderBalanceBefore = await weth.balanceOf(signer.address)
            const receiverBalanceBefore = await weth.balanceOf(receiver.address)

            const transferTx = await weth.connect(signer).transfer(receiver.address, depositAmount)
            const receipt = await transferTx.wait()

            const logs = receipt.logs.map(log => weth.interface.parseLog(log))
            expect(logs.length).to.be.equal(1)
            expect(logs[0].name).to.be.equal('Transfer')
            expect(logs[0].args[0]).to.be.equal(signer.address)
            expect(logs[0].args[1]).to.be.equal(receiver.address)
            expect(logs[0].args[2]).to.be.equal(depositAmount)

            const senderBalanceAfter = await weth.balanceOf(signer.address)
            const receiverBalanceAfter = await weth.balanceOf(receiver.address)

            expect(senderBalanceAfter).to.equal(senderBalanceBefore.sub(depositAmount))
            expect(receiverBalanceAfter).to.equal(receiverBalanceBefore.add(depositAmount));
        });
    });
});