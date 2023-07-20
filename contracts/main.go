package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://localhost:7545")
	if err != nil {
		log.Fatal(err)
	}
	abiBytes, err := ioutil.ReadFile("TestToken.abi")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the ABI
	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress("0x79b168E4d21DF857168ad29c1c74856984e6448A")
	contract, err := NewTestToken(contractAddress, client, contractABI)
	if err != nil {
		log.Fatal(err)
	}
	fromAddress := common.HexToAddress("0xA2f4bc15b5046E72DFf903749D721CFDfC945ed6")
	toAddress := common.HexToAddress("0x7585b2B0b7405e12682CFB4CA66B1A31F3FEA9AB")

	transferAmount := big.NewInt(100)                                  // Reduced transfer amount
	maxBalance, _ := new(big.Int).SetString("1000000000000000000", 10) // Set maximum balance available for the sender's address
	if transferAmount.Cmp(maxBalance) > 0 {
		transferAmount.Set(maxBalance) // Set transfer amount to the maximum balance if it exceeds the available balance
	}

	initialBalance, err := contract.BalanceOf(nil, fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initial Balance: %s\n", initialBalance.String())

	// Transfer tokens
	fmt.Println(fromAddress, toAddress, transferAmount)
	err = contract.Transfer(nil, fromAddress, toAddress, transferAmount)
	if err != nil {
		log.Fatal(err)
	}

	// Check final balance
	finalBalance, err := contract.BalanceOf(nil, fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Final Balance: %s\n", finalBalance.String())
}

func NewTestToken(address common.Address, client *ethclient.Client, contractABI abi.ABI) (*TestToken, error) {
	return &TestToken{
		Address: address,
		Client:  client,
		ABI:     contractABI,
	}, nil
}

type TestToken struct {
	Address common.Address
	Client  *ethclient.Client
	ABI     abi.ABI
}

func (t *TestToken) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	callData, err := t.ABI.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &t.Address,
		Data: callData,
	}
	result, err := t.Client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	err = t.ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (t *TestToken) Transfer(ctx context.Context, sender common.Address, receiver common.Address, amount *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check sender's balance
	initialSenderBalance, err := t.BalanceOf(ctx, sender)
	if err != nil {
		return err
	}
	fmt.Println("From", sender)
	fmt.Printf("Sender's Initial Balance: %s\n", initialSenderBalance.String())

	// Check receiver's initial balance
	initialReceiverBalance, err := t.BalanceOf(ctx, receiver)
	if err != nil {
		return err
	}
	fmt.Println("To", receiver)
	fmt.Printf("Receiver's Initial Balance: %s\n", initialReceiverBalance.String())

	// Check if sender has sufficient balance
	if initialSenderBalance.Cmp(amount) < 0 {
		return fmt.Errorf("Insufficient balance")
	}

	// Prepare the data for the transfer function
	data, err := t.ABI.Pack("transfer", receiver, amount)
	if err != nil {
		return err
	}
	fmt.Println("Amt", amount)

	// Get the private key of the sender account
	privateKey, err := crypto.HexToECDSA("5be20b38c52b8557315322bbbf5347fb5425187e863693cc85106b1eeb083431")
	if err != nil {
		return err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := t.Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}
	gasPrice, err := t.Client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	// Assuming the gas limit for the transfer function is 21000
	gasLimit := uint64(100000)

	// Create the transaction
	tx := types.NewTransaction(nonce, t.Address, big.NewInt(0), gasLimit, gasPrice, data)
	// Sign the transaction
	chainID, err := t.Client.ChainID(ctx)
	if err != nil {
		return err
	}
	fmt.Println(chainID)
	// chainID := big.NewInt(5)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}
	fmt.Println(signedTx)
	// Send the transaction
	err = t.Client.SendTransaction(ctx, signedTx)
	if err != nil {

		return err
	}

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(ctx, t.Client, signedTx)
	if err != nil {
		return err
	}
	fmt.Println("Recipet", receipt)
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transfer failed: transaction status is unsuccessful")
	}

	finalBalance, err := t.BalanceOf(ctx, sender)
	if err != nil {
		return err
	}
	fmt.Printf("Sender's Final Balance: %s\n", finalBalance.String())

	// Check receiver's final balance
	finalReceiverBalance, err := t.BalanceOf(ctx, receiver)
	if err != nil {
		return err
	}
	fmt.Printf("Receiver's Final Balance: %s\n", finalReceiverBalance.String())

	// Compare balances
	if initialSenderBalance.Sub(initialSenderBalance, amount).Cmp(finalBalance) != 0 {
		return fmt.Errorf("Transfer failed: Sender's balance not deducted correctly")
	}

	if initialReceiverBalance.Add(initialReceiverBalance, amount).Cmp(finalReceiverBalance) != 0 {
		return fmt.Errorf("Transfer failed: Receiver's balance not added correctly")
	}

	return nil
}
