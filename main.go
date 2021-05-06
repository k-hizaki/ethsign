package main

import (
	"fmt"
	"log"
	"strings"
	"math/big"
	"context"
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

func main() {
	// generate a private key
	keyA, _ := crypto.GenerateKey()
	keyB, _ := crypto.GenerateKey()

	// TransactOpts c.f. https://github.com/ethereum/go-ethereum/blob/0e00ee42ec4e43ce3b9b1ffdadea3c66aa6eeba4/accounts/abi/bind/base.go#L47
	authA := bind.NewKeyedTransactor(keyA)
	authB := bind.NewKeyedTransactor(keyB)
	
	// prepare the data for the simulated ethereum blockchain
	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei
	addressA := authA.From
	addressB := authB.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		addressA: {
			Balance: balance,
		},
		addressB: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(4712388)
	
	// get the simulated blockchain
	sim := backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)

	// -------- Done in offline terminal --------
	// get the transaction signed by keyA
	// parse the abi data of intstorage
	parsedAbi, err := abi.JSON(strings.NewReader(IntStorageABI))
	if err != nil {
		log.Fatalf("Failed to parse the abi")
	}
	authA.NoSend = true // sign only option, no actual transaction sending
	_, signedTxOffline, _, err := bind.DeployContract(authA, parsedAbi, common.FromHex(IntStorageBin), sim)
	if err != nil {
		log.Fatalf("Failed to get signed transaction")
	}

	// serialize the signed transaction
	var buf bytes.Buffer
	signedTxOffline.EncodeRLP(&buf)
	fmt.Printf("%x\n", buf)


	// -------- Done in online terminal --------
	// send transaction from authB
	var tx types.Transaction
	reader := bytes.NewReader(buf.Bytes())
	stream := rlp.NewStream(reader, uint64(reader.Size()))
	tx.DecodeRLP(stream)

	// send the transaction
	err = sim.SendTransaction(ensureContext(authB.Context), &tx)

	// mine new block
	sim.Commit()

	b, err := sim.BalanceAt(authB.Context, addressA, big.NewInt(0))
	fmt.Printf("balance of A at block level 0: %swei\n", b.String())
	b, err = sim.BalanceAt(authB.Context, addressA, big.NewInt(1))
	fmt.Printf("balance of A at block level 1: %swei\n", b.String())
	b, err = sim.BalanceAt(authB.Context, addressA, big.NewInt(2))
	fmt.Printf("balance of A at block level 2: %swei\n", b.String())

	b, err = sim.BalanceAt(authB.Context, addressB, big.NewInt(0))
	fmt.Printf("balance of B at block level 0: %swei\n", b.String())
	b, err = sim.BalanceAt(authB.Context, addressB, big.NewInt(1))
	fmt.Printf("balance of B at block level 1: %swei\n", b.String())
	b, err = sim.BalanceAt(authB.Context, addressB, big.NewInt(2))
	fmt.Printf("balance of B at block level 2: %swei\n", b.String())
}

func ensureContext(ctx context.Context) (context.Context) {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}

