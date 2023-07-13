// SPDX-License-Identifier: MIT 
pragma solidity >=0.4.22 <0.9.0;

contract TestToken {
    string public name;
    string public symbol;
    uint256 public totalSupply;

    mapping(address => uint256) balances;
    mapping(address => mapping(address => uint256)) allowed;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    constructor(uint256 _initialSupply) public {
        name = "Test Token";
        symbol = "TTK";
        totalSupply = _initialSupply;
        balances[msg.sender] = _initialSupply;
        assert(totalSupply > 0 && totalSupply >= _initialSupply);
        assert(balances[msg.sender] == _initialSupply);
    }

    function balanceOf(address _owner) public view returns (uint256) {
        return balances[_owner];
    }
   
    function transfer(address _to, uint256 _value) public returns (bool) {
        require(_to != address(0), "Invalid recipient address");
        require(_value <= balances[msg.sender], "Insufficient balance");

        balances[msg.sender] -= _value;
        balances[_to] += _value;
        emit Transfer(msg.sender, _to, _value);
        return true;
    }
    
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool) {
        require(_to != address(0), "Invalid recipient address");
        require(_value <= balances[_from], "Insufficient balance");
        require(_value <= allowed[_from][msg.sender], "Insufficient allowance");

        balances[_from] -= _value;
        balances[_to] += _value;
        allowed[_from][msg.sender] -= _value;
        emit Transfer(_from, _to, _value);
        return true;
    }
    
    function approve(address _spender, uint256 _value) public returns (bool) {
        require(_spender != address(0), "Invalid spender address");
        require(_value <= balances[msg.sender], "Insufficient balance");

        allowed[msg.sender][_spender] = _value;
        balances[msg.sender] -= _value; // Deduct the allowance from the sender's balance
        emit Approval(msg.sender, _spender, _value);
        return true;
    }

    function allowance(address _owner, address _spender) public view returns (uint256) {
        return allowed[_owner][_spender];
    }
}
