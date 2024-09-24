const { expect } = require("chai");
const { ethers } = require("hardhat");
const { abi: FACTORY_ABI, bytecode: FACTORY_BYTECODE } = require('@uniswap/v3-core/artifacts/contracts/UniswapV3Factory.sol/UniswapV3Factory.json');
const { abi: POOL_ABI } = require('@uniswap/v3-core/artifacts/contracts/UniswapV3Pool.sol/UniswapV3Pool.json');
const { abi: NFT_MANAGER_ABI, bytecode: NFT_MANAGER_BYTECODE } = require('@uniswap/v3-periphery/artifacts/contracts/NonfungiblePositionManager.sol/NonfungiblePositionManager.json');
const { abi: ROUTER_ABI, bytecode: ROUTER_BYTECODE } = require('@uniswap/v3-periphery/artifacts/contracts/SwapRouter.sol/SwapRouter.json');

const provider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider)
const receiver = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", provider)

const FEE_TIER = 3000; // 0.3% fee tier
const INITIAL_LIQUIDITY = ethers.utils.parseEther("10");
const SWAP_AMOUNT = ethers.utils.parseEther("1")

describe("Uniswap V3", function() {
    let weth, factory, router, nftManager, erc20First, erc20Second

    before(async () => {
        // Deploy WETH
        const WETH9 = await ethers.getContractFactory("WETH9", signer);
        weth = await WETH9.deploy();
        await weth.deployed();

        // Deploy Factory
        const Factory = new ethers.ContractFactory(FACTORY_ABI, FACTORY_BYTECODE, signer);
        factory = await Factory.deploy();
        await factory.deployed();

        // Deploy Router
        const Router = new ethers.ContractFactory(ROUTER_ABI, ROUTER_BYTECODE, signer);
        router = await Router.deploy(factory.address, weth.address);
        await router.deployed();

        // Deploy NFT Position Manager
        const NFTManager = new ethers.ContractFactory(NFT_MANAGER_ABI, NFT_MANAGER_BYTECODE, signer);
        nftManager = await NFTManager.deploy(factory.address, weth.address, ethers.constants.AddressZero);
        await nftManager.deployed();

        // Deploy token contracts
        const Token = await ethers.getContractFactory("Token", signer);
        erc20First = await Token.deploy("First", "F");
        await erc20First.deployed();
        erc20Second = await Token.deploy("Second", "S");
        await erc20Second.deployed();
    })

    it("should create a new pool", async function() {
        // Ensure tokens are in the correct order (lower address first)
        const [token0Address, token1Address] = erc20First.address.toLowerCase() < erc20Second.address.toLowerCase()
            ? [erc20First.address, erc20Second.address]
            : [erc20Second.address, erc20First.address];

        // Create the pool
        const tx = await factory.createPool(token0Address, token1Address, FEE_TIER)
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => factory.interface.parseLog(log))

        expect(logs[0].name).to.be.equal("PoolCreated")
        expect(logs[0].args[0]).to.be.equal(token0Address)
        expect(logs[0].args[1]).to.be.equal(token1Address)
        expect(logs[0].args[2]).to.be.equal(FEE_TIER)

        // Get the pool address
        const poolAddress = await factory.getPool(token0Address, token1Address, FEE_TIER);
        expect(poolAddress).to.not.equal(ethers.constants.AddressZero);
    });

    it("should initialize the pool", async function() {
        const [token0Address, token1Address] = erc20First.address.toLowerCase() < erc20Second.address.toLowerCase()
            ? [erc20First.address, erc20Second.address]
            : [erc20Second.address, erc20First.address];

        const poolAddress = await factory.getPool(token0Address, token1Address, FEE_TIER);
        const pool = await ethers.getContractAt(POOL_ABI, poolAddress, signer);

        // Set the initial price (1:1 in this case)
        const initialSqrtPrice = ethers.BigNumber.from("79228162514264337593543950336");
        const tx = await pool.initialize(initialSqrtPrice);
        await tx.wait()

        // Verify the pool is initialized
        const slot0 = await pool.slot0();
        expect(slot0.sqrtPriceX96).to.equal(initialSqrtPrice);
    });

    it("should add liquidity to the pool", async function() {
        const [token0Address, token1Address] = erc20First.address.toLowerCase() < erc20Second.address.toLowerCase()
            ? [erc20First.address, erc20Second.address]
            : [erc20Second.address, erc20First.address];

        // Approve tokens for NFT manager
        const tx1 = await erc20First.connect(signer).approve(nftManager.address, INITIAL_LIQUIDITY);
        await tx1.wait()

        const tx2 = await erc20Second.connect(signer).approve(nftManager.address, INITIAL_LIQUIDITY);
        await tx2.wait()

        // Calculate min and max tick for full range
        const minTick = getMinTick(FEE_TIER);
        const maxTick = getMaxTick(FEE_TIER);

        // Add liquidity
        const tx = await nftManager.connect(signer).mint({
            token0: token0Address,
            token1: token1Address,
            fee: FEE_TIER,
            tickLower: minTick,
            tickUpper: maxTick,
            amount0Desired: INITIAL_LIQUIDITY,
            amount1Desired: INITIAL_LIQUIDITY,
            amount0Min: 0,
            amount1Min: 0,
            recipient: signer.address,
            deadline: Math.floor(Date.now() / 1000) + 300
        });
        const receipt = await tx.wait()

        // Get the tokenId of the newly minted position
        const mintEvent = receipt.events.find(event => event.event === 'IncreaseLiquidity');
        const tokenId = mintEvent.args.tokenId;

        // Fetch position info
        const position = await nftManager.positions(tokenId);

        // Check that liquidity has been added
        expect(position.liquidity).to.be.gt(0);
    });


    it("should swap tokens", async function() {
        // User approves router to spend tokens
        const approveTx = await erc20First.approve(router.address, SWAP_AMOUNT);
        await approveTx.wait()

        // Record balances before swap
        const token0BalanceBefore = await erc20First.balanceOf(signer.address);
        const token1BalanceBefore = await erc20Second.balanceOf(signer.address);

        // Perform swap
        const swapParams = {
            tokenIn: erc20First.address,
            tokenOut: erc20Second.address,
            fee: FEE_TIER,
            recipient: signer.address,
            deadline: Math.floor(Date.now() / 1000) + 300,
            amountIn: SWAP_AMOUNT,
            amountOutMinimum: 0,
            sqrtPriceLimitX96: 0
        }

        const tx = await router.exactInputSingle(swapParams);
        await tx.wait()

        // Check balances after swap
        const token0BalanceAfter = await erc20First.balanceOf(signer.address);
        const token1BalanceAfter = await erc20Second.balanceOf(signer.address);

        expect(token0BalanceAfter).to.be.lt(token0BalanceBefore);
        expect(token1BalanceAfter).to.be.gt(token1BalanceBefore);
    });

    it('create pool for WETH / ERC20 and swap SWRT to ERC20 using Router', async () => {
        // Create the WETH pool
        const createPoolTx = await factory.createPool(weth.address, erc20First.address, FEE_TIER);
        await createPoolTx.wait()
        const poolAddress = await factory.getPool(weth.address, erc20First.address, FEE_TIER);
        const pool = await ethers.getContractAt(POOL_ABI, poolAddress, signer);

        // Initialize pool
        const initTx = await pool.initialize(ethers.utils.parseUnits("1", 18)); // Initial price 1 WETH = 1 ERC20
        await initTx.wait()

        // Put liquidity to the pool
        const depositTx = await weth.deposit({ value: ethers.utils.parseEther("1") });
        await depositTx.wait()
        const approveWETHTx = await weth.approve(nftManager.address, ethers.utils.parseEther("1"));
        await approveWETHTx.wait()
        const approveERCTx = await erc20First.approve(nftManager.address, ethers.utils.parseEther("1"));
        await approveERCTx.wait()

        // Mint position
        console.log('before mint')
        const mintTx = await nftManager.mint({
            token0: weth.address,
            token1: erc20First.address,
            fee: FEE_TIER,
            tickLower: -100,
            tickUpper: 200,
            amount0Desired: ethers.utils.parseEther("1"),
            amount1Desired: ethers.utils.parseEther("1"),
            amount0Min: 0,
            amount1Min: 0,
            recipient: signer.address,
            deadline: Math.floor(Date.now() / 1000) + 3600
        });
        await mintTx.wait()
        console.log('after mint')

        // Execute swap
        const initialERC20Balance = await erc20First.balanceOf(signer.address);

        // Correctly define the path
        const path = ethers.utils.solidityPack(
            ['address', 'uint24', 'address'],
            [weth.address, FEE_TIER, erc20First.address]
        );
        const swapAmount = ethers.utils.parseEther("0.1");
        console.log('before swap')
        const executeTx = await router.exactInput(
            {
                recipient: signer.address,
                deadline: Math.floor(Date.now() / 1000) + 36000,
                amountIn: swapAmount,
                amountOutMinimum: 0,
                path: path,
            },
            { value: swapAmount }
        );
        await executeTx.wait()
        console.log('after swap')
        const finalERC20Balance = await erc20First.balanceOf(signer.address);
        expect(finalERC20Balance.sub(initialERC20Balance)).to.be.gt(0);
        console.log("ERC20 tokens received:", ethers.utils.formatEther(finalERC20Balance.sub(initialERC20Balance)));
    })
})


// helpers/encodePriceSqrt.js
function encodePriceSqrt(reserve1, reserve0) {
    return ethers.BigNumber.from(
        new bn(reserve1.toString())
            .div(reserve0.toString())
            .sqrt()
            .multipliedBy(new bn(2).pow(96))
            .integerValue(3)
            .toString()
    )
}

// helpers/ticks.js
const TICK_SPACINGS = {
    500: 10,
    3000: 60,
    10000: 200
};

function getMinTick(tickSpacing) {
    return Math.ceil(-887272 / tickSpacing) * tickSpacing;
}

function getMaxTick(tickSpacing) {
    return Math.floor(887272 / tickSpacing) * tickSpacing;
}