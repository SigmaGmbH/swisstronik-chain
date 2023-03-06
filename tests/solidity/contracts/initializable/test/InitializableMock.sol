pragma solidity ^0.8.0;

import "../Initializable.sol";
import "../Petrifiable.sol";


contract LifecycleMock is Initializable, Petrifiable {
    function initializeMock() public {
        initialized();
    }

    function petrifyMock() public {
        petrify();
    }
}
