// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.6.0;

// Minimal Uniswap V2 style AMM for testing ProbeChain DEX functionality
// Combines Factory+Pair+Router into one contract for simplicity

interface IERC20 {
    function balanceOf(address) external view returns (uint);
    function transfer(address to, uint amount) external returns (bool);
    function transferFrom(address from, address to, uint amount) external returns (bool);
    function approve(address spender, uint amount) external returns (bool);
}

contract MiniSwap {
    address public owner;
    address public WPROBE;

    struct Pair {
        address token0;
        address token1;
        uint reserve0;
        uint reserve1;
        uint totalLP;
        mapping(address => uint) lpBalance;
    }

    uint public pairCount;
    mapping(uint => Pair) public pairs;
    mapping(address => mapping(address => uint)) public getPairId;

    event PairCreated(uint indexed pairId, address token0, address token1);
    event Swap(uint indexed pairId, address indexed sender, uint amountIn, uint amountOut, bool zeroForOne);
    event AddLiquidity(uint indexed pairId, address indexed sender, uint amount0, uint amount1, uint lp);
    event RemoveLiquidity(uint indexed pairId, address indexed sender, uint amount0, uint amount1, uint lp);

    constructor(address _wprobe) public {
        owner = msg.sender;
        WPROBE = _wprobe;
        pairCount = 0;
    }

    function createPair(address tokenA, address tokenB) external returns (uint pairId) {
        require(tokenA != tokenB, "identical");
        (address t0, address t1) = tokenA < tokenB ? (tokenA, tokenB) : (tokenB, tokenA);
        require(getPairId[t0][t1] == 0, "exists");
        pairCount++;
        pairId = pairCount;
        pairs[pairId].token0 = t0;
        pairs[pairId].token1 = t1;
        getPairId[t0][t1] = pairId;
        getPairId[t1][t0] = pairId;
        emit PairCreated(pairId, t0, t1);
    }

    function addLiquidity(uint pairId, uint amount0, uint amount1) external returns (uint lp) {
        Pair storage p = pairs[pairId];
        require(p.token0 != address(0), "no pair");
        IERC20(p.token0).transferFrom(msg.sender, address(this), amount0);
        IERC20(p.token1).transferFrom(msg.sender, address(this), amount1);
        if (p.totalLP == 0) {
            lp = sqrt(amount0 * amount1);
        } else {
            uint lp0 = amount0 * p.totalLP / p.reserve0;
            uint lp1 = amount1 * p.totalLP / p.reserve1;
            lp = lp0 < lp1 ? lp0 : lp1;
        }
        require(lp > 0, "no lp");
        p.reserve0 += amount0;
        p.reserve1 += amount1;
        p.totalLP += lp;
        p.lpBalance[msg.sender] += lp;
        emit AddLiquidity(pairId, msg.sender, amount0, amount1, lp);
    }

    function removeLiquidity(uint pairId, uint lp) external returns (uint amount0, uint amount1) {
        Pair storage p = pairs[pairId];
        require(p.lpBalance[msg.sender] >= lp, "low lp");
        amount0 = lp * p.reserve0 / p.totalLP;
        amount1 = lp * p.reserve1 / p.totalLP;
        p.lpBalance[msg.sender] -= lp;
        p.totalLP -= lp;
        p.reserve0 -= amount0;
        p.reserve1 -= amount1;
        IERC20(p.token0).transfer(msg.sender, amount0);
        IERC20(p.token1).transfer(msg.sender, amount1);
        emit RemoveLiquidity(pairId, msg.sender, amount0, amount1, lp);
    }

    function swap(uint pairId, address tokenIn, uint amountIn, uint amountOutMin) external returns (uint amountOut) {
        Pair storage p = pairs[pairId];
        bool zeroForOne = tokenIn == p.token0;
        require(zeroForOne || tokenIn == p.token1, "bad token");
        (uint rIn, uint rOut) = zeroForOne ? (p.reserve0, p.reserve1) : (p.reserve1, p.reserve0);
        IERC20(tokenIn).transferFrom(msg.sender, address(this), amountIn);
        // x * y = k, with 0.3% fee
        uint amountInWithFee = amountIn * 997;
        amountOut = amountInWithFee * rOut / (rIn * 1000 + amountInWithFee);
        require(amountOut >= amountOutMin, "slippage");
        require(amountOut > 0, "zero out");
        address tokenOut = zeroForOne ? p.token1 : p.token0;
        IERC20(tokenOut).transfer(msg.sender, amountOut);
        if (zeroForOne) { p.reserve0 += amountIn; p.reserve1 -= amountOut; }
        else { p.reserve1 += amountIn; p.reserve0 -= amountOut; }
        emit Swap(pairId, msg.sender, amountIn, amountOut, zeroForOne);
    }

    function getReserves(uint pairId) external view returns (uint r0, uint r1, uint lp) {
        Pair storage p = pairs[pairId];
        return (p.reserve0, p.reserve1, p.totalLP);
    }

    function getAmountOut(uint amountIn, uint rIn, uint rOut) external pure returns (uint) {
        uint aif = amountIn * 997;
        return aif * rOut / (rIn * 1000 + aif);
    }

    function lpBalanceOf(uint pairId, address user) external view returns (uint) {
        return pairs[pairId].lpBalance[user];
    }

    function sqrt(uint x) internal pure returns (uint y) {
        if (x == 0) return 0;
        uint z = (x + 1) / 2; y = x;
        while (z < y) { y = z; z = (x / z + z) / 2; }
    }
}
