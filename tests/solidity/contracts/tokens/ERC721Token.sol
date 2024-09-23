pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

contract ERC721Token is ERC721URIStorage {
    uint256 private _tokenIdCounter;

    constructor(string memory name, string memory symbol) ERC721(name, symbol) {}

    function createItem(address player, string memory tokenURI) public returns (uint256) {
        uint256 newItemId = _tokenIdCounter;
        _safeMint(player, newItemId);
        _setTokenURI(newItemId, tokenURI);

        _tokenIdCounter += 1;
        return newItemId;
    }
}
