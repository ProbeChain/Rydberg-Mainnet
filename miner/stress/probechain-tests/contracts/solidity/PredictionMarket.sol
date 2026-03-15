pragma solidity ^0.6.0;
contract PredictionMarket {
    address public owner;
    uint public marketCount;
    struct Market {
        string question;
        uint totalYes;
        uint totalNo;
        bool resolved;
        bool outcome;
        mapping(address => uint) yesShares;
        mapping(address => uint) noShares;
        mapping(address => bool) claimed;
    }
    mapping(uint => Market) public markets;
    event MarketCreated(uint indexed id, string question);
    event SharesBought(uint indexed id, address buyer, bool isYes, uint amount);
    event MarketResolved(uint indexed id, bool outcome);
    event WinningsClaimed(uint indexed id, address claimer, uint amount);
    constructor() public { owner = msg.sender; }
    function createMarket(string calldata q) external returns (uint id) {
        id = ++marketCount;
        markets[id].question = q;
        emit MarketCreated(id, q);
    }
    function buyYes(uint id) external payable { require(!markets[id].resolved && id<=marketCount && id>0); markets[id].yesShares[msg.sender] += msg.value; markets[id].totalYes += msg.value; emit SharesBought(id,msg.sender,true,msg.value); }
    function buyNo(uint id) external payable { require(!markets[id].resolved && id<=marketCount && id>0); markets[id].noShares[msg.sender] += msg.value; markets[id].totalNo += msg.value; emit SharesBought(id,msg.sender,false,msg.value); }
    function resolve(uint id, bool outcome) external { require(msg.sender==owner && !markets[id].resolved); markets[id].resolved = true; markets[id].outcome = outcome; emit MarketResolved(id, outcome); }
    function claim(uint id) external {
        Market storage m = markets[id];
        require(m.resolved && !m.claimed[msg.sender]);
        m.claimed[msg.sender] = true;
        uint shares = m.outcome ? m.yesShares[msg.sender] : m.noShares[msg.sender];
        uint total = m.outcome ? m.totalYes : m.totalNo;
        uint pool = m.totalYes + m.totalNo;
        if (shares > 0 && total > 0) {
            uint payout = shares * pool / total;
            (bool ok,) = msg.sender.call{value:payout}("");
            require(ok);
            emit WinningsClaimed(id, msg.sender, payout);
        }
    }
    function getShares(uint id, address u) external view returns (uint yes, uint no) { return (markets[id].yesShares[u], markets[id].noShares[u]); }
    function getMarketInfo(uint id) external view returns (uint totalYes, uint totalNo, bool resolved, bool outcome) { Market storage m=markets[id]; return(m.totalYes,m.totalNo,m.resolved,m.outcome); }
}
