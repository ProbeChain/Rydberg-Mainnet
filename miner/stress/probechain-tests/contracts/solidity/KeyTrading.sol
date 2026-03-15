// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.6.0;

contract KeyTrading {
    mapping(address => uint256) public sharesSupply;
    mapping(address => mapping(address => uint256)) public sharesBalance;
    address public protocolFeeDestination;
    uint256 public protocolFeePercent;
    uint256 public subjectFeePercent;
    address public owner;

    event Trade(address indexed trader, address indexed subject, bool isBuy, uint256 amount, uint256 price, uint256 supply);

    constructor() public {
        owner = msg.sender;
        protocolFeeDestination = msg.sender;
        protocolFeePercent = 250;
        subjectFeePercent = 250;
    }

    function _getPrice(uint256 supply, uint256 amount) internal pure returns (uint256) {
        uint256 sum1 = supply == 0 ? 0 : (supply - 1) * supply * (2 * (supply - 1) + 1) / 6;
        uint256 sum2 = (supply + amount - 1) * (supply + amount) * (2 * (supply + amount - 1) + 1) / 6;
        return (sum2 - sum1) * 1 ether / 16000;
    }

    function getBuyPrice(address s, uint256 a) public view returns (uint256) { return _getPrice(sharesSupply[s], a); }
    function getSellPrice(address s, uint256 a) public view returns (uint256) {
        require(sharesSupply[s] >= a, "low supply");
        return _getPrice(sharesSupply[s] - a, a);
    }
    function getBuyPriceAfterFee(address s, uint256 a) public view returns (uint256) {
        uint256 p = getBuyPrice(s, a);
        return p + p * protocolFeePercent / 10000 + p * subjectFeePercent / 10000;
    }
    function getSellPriceAfterFee(address s, uint256 a) public view returns (uint256) {
        uint256 p = getSellPrice(s, a);
        return p - p * protocolFeePercent / 10000 - p * subjectFeePercent / 10000;
    }

    function buyShares(address subject, uint256 amount) external payable {
        uint256 supply = sharesSupply[subject];
        require(supply > 0 || subject == msg.sender, "first share self-buy");
        uint256 price = _getPrice(supply, amount);
        uint256 pFee = price * protocolFeePercent / 10000;
        uint256 sFee = price * subjectFeePercent / 10000;
        require(msg.value >= price + pFee + sFee, "insufficient payment");
        sharesBalance[subject][msg.sender] += amount;
        sharesSupply[subject] = supply + amount;
        emit Trade(msg.sender, subject, true, amount, price, supply + amount);
        _send(protocolFeeDestination, pFee);
        _send(subject, sFee);
        uint256 excess = msg.value - price - pFee - sFee;
        if (excess > 0) _send(msg.sender, excess);
    }

    function sellShares(address subject, uint256 amount) external {
        uint256 supply = sharesSupply[subject];
        require(supply > amount, "cant sell last");
        require(sharesBalance[subject][msg.sender] >= amount, "insufficient shares");
        uint256 price = _getPrice(supply - amount, amount);
        uint256 pFee = price * protocolFeePercent / 10000;
        uint256 sFee = price * subjectFeePercent / 10000;
        sharesBalance[subject][msg.sender] -= amount;
        sharesSupply[subject] = supply - amount;
        emit Trade(msg.sender, subject, false, amount, price, supply - amount);
        _send(msg.sender, price - pFee - sFee);
        _send(protocolFeeDestination, pFee);
        _send(subject, sFee);
    }

    function setFeeDestination(address d) external { require(msg.sender == owner); protocolFeeDestination = d; }
    function setProtocolFeePercent(uint256 f) external { require(msg.sender == owner); require(f <= 500); protocolFeePercent = f; }
    function setSubjectFeePercent(uint256 f) external { require(msg.sender == owner); require(f <= 500); subjectFeePercent = f; }

    function _send(address to, uint256 amt) internal {
        if (amt == 0) return;
        (bool ok,) = to.call{value: amt}("");
        require(ok, "transfer failed");
    }
}
