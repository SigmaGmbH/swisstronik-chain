pragma solidity ^0.8;

contract BlockInfo {
    function getBlockBaseFee() public view returns (uint256) {
        return block.basefee;
    }

    function getBlockChainId() public view returns (uint256) {
        return block.chainid;
    }

    function getBlockCoinbase() public view returns (address) {
        return block.coinbase;
    }

    function getBlockDifficulty() public view returns (uint256) {
        return block.difficulty;
    }

    function getBlockGaslimit() public view returns (uint256) {
        return block.gaslimit;
    }

    function getBlockNumber() public view returns (uint256) {
        return block.number;
    }

    function getBlockTimestamp() public view returns (uint256) {
        return block.timestamp;
    }

    // block.blockhash is a special case as it requires a block number as input
    function getBlockhash(uint256 blockNumber) public view returns (bytes32) {
        return blockhash(blockNumber);
    }

    // Getter for all block information in one call
    function getAllBlockInfo() public view returns (
        uint256 basefee,
        uint256 chainid,
        address coinbase,
        uint256 difficulty,
        uint256 gaslimit,
        uint256 number,
        uint256 timestamp
    ) {
        return (
            block.basefee,
            block.chainid,
            block.coinbase,
            block.difficulty,
            block.gaslimit,
            block.number,
            block.timestamp
        );
    }
}