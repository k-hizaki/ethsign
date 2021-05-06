// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.4.20;

contract IntStorage {
  uint public intdata;

  function set(uint x) public {
    intdata = x;
  }

  function get() public view returns (uint) {
    return intdata;
  }
}
