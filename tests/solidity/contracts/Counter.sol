pragma solidity ^0.8;

contract Counter {
  uint256 public counter = 0;
  string internal constant ERROR_TOO_LOW = "COUNTER_TOO_LOW";

  event Changed(uint256 counter);
  event Added(uint256 counter);

  function add() public {
    counter++;
    emit Added(counter);
    emit Changed(counter);
  }

  function subtract() public {
    require(counter > 0, ERROR_TOO_LOW);
    counter--;
    emit Changed(counter);
  }
}
