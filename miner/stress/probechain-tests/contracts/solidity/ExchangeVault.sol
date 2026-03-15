pragma solidity ^0.6.0;
contract ExchangeVault {
    address public owner;
    mapping(address => uint) public deposits;
    uint public totalDeposits;
    event Deposited(address indexed user, uint amount);
    event Withdrawn(address indexed user, uint amount);
    event BatchSettled(bytes32 indexed settleHash, uint count);
    constructor() public { owner = msg.sender; }
    function deposit() external payable { deposits[msg.sender] += msg.value; totalDeposits += msg.value; emit Deposited(msg.sender, msg.value); }
    function withdraw(uint amount) external { require(deposits[msg.sender] >= amount); deposits[msg.sender] -= amount; totalDeposits -= amount; (bool ok,) = msg.sender.call{value:amount}(""); require(ok); emit Withdrawn(msg.sender, amount); }
    function batchSettle(address[] calldata users, uint[] calldata amounts, bytes32 settleHash) external {
        require(msg.sender == owner && users.length == amounts.length);
        for(uint i=0; i<users.length; i++) { deposits[users[i]] += amounts[i]; totalDeposits += amounts[i]; }
        emit BatchSettled(settleHash, users.length);
    }
    function getDeposit(address u) external view returns (uint) { return deposits[u]; }
}
