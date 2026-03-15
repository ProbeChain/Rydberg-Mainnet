// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.6.0;

contract IdentityRegistry {
    uint256 private _nextTokenId;
    address public owner;
    
    struct Agent { address registrant; string uri; }
    mapping(uint256 => Agent) private _agents;
    mapping(uint256 => address) private _approvals;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => bool)) private _operatorApprovals;
    
    event Registered(uint256 indexed agentId, string agentURI, address indexed registrant);
    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
    event Approval(address indexed tokenOwner, address indexed approved, uint256 indexed tokenId);
    event ApprovalForAll(address indexed tokenOwner, address indexed operator, bool approved);
    
    constructor() public { owner = msg.sender; _nextTokenId = 1; }
    
    function register(string calldata agentURI) external returns (uint256 agentId) {
        agentId = _nextTokenId;
        _nextTokenId = _nextTokenId + 1;
        _agents[agentId].registrant = msg.sender;
        _agents[agentId].uri = agentURI;
        _balances[msg.sender] = _balances[msg.sender] + 1;
        emit Registered(agentId, agentURI, msg.sender);
        emit Transfer(address(0), msg.sender, agentId);
    }
    function getAgentWallet(uint256 agentId) external view returns (address) {
        require(_agents[agentId].registrant != address(0), "not exist");
        return _agents[agentId].registrant;
    }
    function tokenURI(uint256 agentId) external view returns (string memory) {
        require(_agents[agentId].registrant != address(0), "not exist");
        return _agents[agentId].uri;
    }
    function totalAgents() external view returns (uint256) { return _nextTokenId - 1; }
    function ownerOf(uint256 tokenId) public view returns (address) {
        address a = _agents[tokenId].registrant;
        require(a != address(0), "not exist");
        return a;
    }
    function balanceOf(address addr) external view returns (uint256) { return _balances[addr]; }
    function approve(address to, uint256 tokenId) external {
        require(msg.sender == ownerOf(tokenId), "not owner");
        _approvals[tokenId] = to;
        emit Approval(msg.sender, to, tokenId);
    }
    function getApproved(uint256 tokenId) external view returns (address) { return _approvals[tokenId]; }
    function setApprovalForAll(address operator, bool approved) external {
        _operatorApprovals[msg.sender][operator] = approved;
        emit ApprovalForAll(msg.sender, operator, approved);
    }
    function isApprovedForAll(address addr, address operator) external view returns (bool) {
        return _operatorApprovals[addr][operator];
    }
    function transferFrom(address from, address to, uint256 tokenId) external {
        require(ownerOf(tokenId) == from, "wrong from");
        require(msg.sender == from || msg.sender == _approvals[tokenId] || _operatorApprovals[from][msg.sender], "not auth");
        _agents[tokenId].registrant = to;
        _balances[from] = _balances[from] - 1;
        _balances[to] = _balances[to] + 1;
        delete _approvals[tokenId];
        emit Transfer(from, to, tokenId);
    }
}
