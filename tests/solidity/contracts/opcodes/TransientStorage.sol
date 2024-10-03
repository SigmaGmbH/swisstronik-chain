pragma solidity 0.8.27;

contract MulService {
    function setMultiplier(uint multiplier) external {
        assembly {
            tstore(0, multiplier)
        }
    }

    function getMultiplier() private view returns (uint multiplier) {
        assembly {
            multiplier := tload(0)
        }
    }

    function multiply(uint value) external view returns (uint) {
        return value * getMultiplier();
    }
}

contract MulCaller {
    MulService public mulService;

    constructor(MulService _mulService) {
        mulService = _mulService;
    }

    function runMultiply(uint multiplier, uint value) public returns (uint) {
        mulService.setMultiplier(multiplier);
        return mulService.multiply(value);
    }
}