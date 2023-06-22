# Example of PERC20 (Private ERC20)

This directory contains example of PERC20 contract and basic test for it.
PERC20 is modified ERC20 OpenZeppelin contract with disabled events 

## Build & Run

Before running install all dependencies by running:
```sh
npm install 
```

Then you can compile contracts by running:
```sh
npm run compile
```

Or run tests by running:
```sh
npm run test
```

<b>NOTE</b>: You should start Swisstronik local node before running tests, since hardhat does not support encrypted transactions. 
Right now we're working on our fork of hardhat with enabled encryption
