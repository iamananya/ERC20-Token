const TestToken = artifacts.require("TestToken");

contract("TestToken", accounts => {
  let testTokenInstance;
  const initialSupply = 1000000;
  const owner = accounts[0];
  const recipient = accounts[1];

  before(async () => {
    testTokenInstance = await TestToken.deployed();
});
it("should initialize the contract with the correct initial supply", async () => {
    const balance = await testTokenInstance.balanceOf(owner);
    assert.equal(balance.toNumber(), initialSupply);
  });

  it("should transfer tokens correctly", async () => {
    const amountToTransfer = 100;
    const initialOwnerBalance = await testTokenInstance.balanceOf(owner);
    const initialRecipientBalance = await testTokenInstance.balanceOf(recipient);

    await testTokenInstance.transfer(recipient, amountToTransfer);

    const newOwnerBalance = await testTokenInstance.balanceOf(owner);
    const newRecipientBalance = await testTokenInstance.balanceOf(recipient);

    assert.equal(newOwnerBalance.toNumber(), initialOwnerBalance.toNumber() - amountToTransfer);
    assert.equal(newRecipientBalance.toNumber(), initialRecipientBalance.toNumber() + amountToTransfer);
  });

  it("should allow token approval and transferFrom", async () => {
    const amountToApprove = 100;
    const amountToTransferFrom = 50;

    await testTokenInstance.approve(accounts[2], amountToApprove, { from: owner });
    await testTokenInstance.transferFrom(owner, recipient, amountToTransferFrom, { from: accounts[2] });

    const ownerBalance = await testTokenInstance.balanceOf(owner);
    const recipientBalance = await testTokenInstance.balanceOf(recipient);
    const allowance = await testTokenInstance.allowance(owner, accounts[2]);

    assert.equal(ownerBalance.toNumber(), initialSupply - amountToTransferFrom);
    assert.equal(recipientBalance.toNumber(), amountToTransferFrom);
    assert.equal(allowance.toNumber(), amountToApprove);

  });

  it("should return the correct allowance", async () => {
    const allowance = await testTokenInstance.allowance(owner, accounts[2]);
    assert.equal(allowance.toNumber(), 0);
  });
});