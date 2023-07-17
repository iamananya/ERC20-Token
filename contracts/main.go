package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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
	contractAddress := common.HexToAddress("0x486Eb4De3050074287E897754347B58bDb5e3595")
	contract, err := NewTestToken(contractAddress, client, contractABI)
	if err != nil {
		log.Fatal(err)
	}

	address := common.HexToAddress("0xA2f4bc15b5046E72DFf903749D721CFDfC945ed6")

	balance, err := contract.BalanceOf(nil, address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Balance: %s\n", balance.String())

	data, err := contractABI.Pack("symbol")
	if err != nil {
		log.Fatal(err)
	}

	callData := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), callData, nil)
	if err != nil {
		log.Fatal(err)
	}

	symbol, err := contractABI.Unpack("symbol", result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Symbol: %s\n", symbol[0].(string))

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
