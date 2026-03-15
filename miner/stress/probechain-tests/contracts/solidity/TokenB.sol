pragma solidity ^0.6.0;
contract TokenB {
    mapping(address => uint256) public balances;
    mapping(address => mapping(address => uint256)) public allowed;
    uint256 public supply;
    address public minter;
    constructor() public { minter=msg.sender; supply=1e24; balances[msg.sender]=1e24; }
    function totalSupply() public view returns (uint256) { return supply; }
    function balanceOf(address a) public view returns (uint256) { return balances[a]; }
    function transfer(address to, uint256 amt) public returns (bool) { require(balances[msg.sender]>=amt); balances[msg.sender]-=amt; balances[to]+=amt; return true; }
    function approve(address s, uint256 a) public returns (bool) { allowed[msg.sender][s]=a; return true; }
    function allowance(address o, address s) public view returns (uint256) { return allowed[o][s]; }
    function transferFrom(address f, address t, uint256 a) public returns (bool) { require(balances[f]>=a&&allowed[f][msg.sender]>=a); balances[f]-=a; balances[t]+=a; allowed[f][msg.sender]-=a; return true; }
    function mint(address to, uint256 a) public { require(msg.sender==minter); supply+=a; balances[to]+=a; }
}
