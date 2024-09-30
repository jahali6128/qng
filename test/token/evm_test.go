// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package token

import (
	"context"
	"log"
	"math/big"
	"testing"

	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/test/testcommon"
	"github.com/Qitmeer/qng/testutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestCallErc20Contract(t *testing.T) {
	h, err := testutils.StartMockNode(nil)
	if err != nil {
		t.Error(err)
	}
	defer h.Stop()

	if err != nil {
		t.Fatalf("setup harness failed:%v", err)
	}
	testutils.GenerateBlocks(t, h, 1)
	testutils.AssertBlockOrderHeightTotal(t, h, 2, 2, 1)

	lockTime := int64(20)
	spendAmt := types.Amount{Value: 14000 * types.AtomsPerCoin, Id: types.MEERA}
	txid := testutils.SendSelfMockNode(t, h, spendAmt, &lockTime)
	testutils.GenerateBlocksWaitForTxs(t, h, []string{txid.String()})
	fee := int64(2200)
	txH := testutils.SendExportTxMockNode(t, h, txid.String(), 0, spendAmt.Value-fee)
	if txH == nil {
		t.Fatalf("createExportRawTx failed:%v", err)
	}
	log.Println("send tx", txH.String())
	testutils.GenerateBlocksWaitForTxs(t, h, []string{txH.String()})
	acc := h.GetWalletManager().GetAccountByIdx(0)
	if err != nil {
		t.Fatalf("GetAcctInfo failed:%v", err)
	}
	ba, err := h.GetEvmClient().BalanceAt(context.Background(), acc.EvmAcct.Address, nil)
	if err != nil {
		t.Fatalf("GetBalance failed:%v", err)
	}
	assert.Equal(t, ba, new(big.Int).Mul(big.NewInt(1e10), big.NewInt(spendAmt.Value-fee)))
	txS, err := testutils.DeployContract(h, 0, common.FromHex(testcommon.ERC20Code))
	if err != nil {
		t.Fatal(err)
	}
	log.Println("create contract tx:", txS)
	testutils.GenerateBlocksWaitForTxs(t, h, []string{txS})
	// token addr
	txD, err := h.GetEvmClient().TransactionReceipt(context.Background(), common.HexToHash(txS))
	if err != nil {
		t.Fatal(err)
	}
	if txD == nil {
		t.Fatal("create contract failed")
	}
	assert.Equal(t, txD.Status, uint64(0x1))
	log.Println("new contract address:", txD.ContractAddress)
	tokenCall, err := NewToken(txD.ContractAddress, h.GetEvmClient())
	if err != nil {
		t.Fatal(err)
	}
	ba, err = tokenCall.BalanceOf(&bind.CallOpts{}, acc.EvmAcct.Address)
	if err != nil {
		t.Fatal(err)
	}
	allAmount := int64(30000000)
	assert.Equal(t, ba, big.NewInt(allAmount).Mul(big.NewInt(allAmount), big.NewInt(1e18)))
	authCaller, err := testutils.AuthTrans(h.GetBuilder().Get(0))
	if err != nil {
		t.Fatal(err)
	}
	_, err = h.NewAddress()
	if err != nil {
		t.Fatal(err)
	}
	toAmount := int64(2)
	toMeerAmount := big.NewInt(1e18).Mul(big.NewInt(1e18), big.NewInt(2))
	txs := []string{}
	for i := 0; i < 2; i++ {
		h.NewAddress()
		to := h.GetWalletManager().GetAccountByIdx((i + 1)).EvmAcct.Address
		// send 2 meer
		txid, err := testutils.MeerTransfer(h, 0, to, toMeerAmount)
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, txid)
		log.Println(i, "transfer meer:", txid)
		// send 2 token
		tx, err := tokenCall.Transfer(authCaller, to, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, tx.Hash().String())
		log.Println(i, "transfer tx:", tx.Hash().String())
	}
	testutils.GenerateBlocksWaitForTxs(t, h, txs)
	ba, err = tokenCall.BalanceOf(&bind.CallOpts{}, acc.EvmAcct.Address)
	if err != nil {
		t.Fatal(err)
	}
	allAmount -= toAmount * 2
	assert.Equal(t, ba, big.NewInt(allAmount).Mul(big.NewInt(allAmount), big.NewInt(1e18)))
	txs = []string{}
	for i := 1; i < 3; i++ {
		target := h.GetWalletManager().GetAccountByIdx(i).EvmAcct.Address
		meerBa, err := h.GetEvmClient().BalanceAt(context.Background(), target, nil)
		if err != nil {
			log.Fatal(err)
		}
		assert.Equal(t, meerBa, toMeerAmount)
		ba, err = tokenCall.BalanceOf(&bind.CallOpts{}, target)
		if err != nil {
			t.Fatal(err)
		}
		log.Println(i, "address", target.String(), "balance", ba)
		assert.Equal(t, ba, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
		h.NewAddress()
		authCaller, err := testutils.AuthTrans(h.GetBuilder().Get(i))
		if err != nil {
			t.Fatal(err)
		}
		tx, err := tokenCall.Transfer(authCaller, h.GetWalletManager().GetAccountByIdx(i+2).EvmAcct.Address, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
		if err != nil {
			t.Fatal(err)
		}
		log.Println(i, "transfer tx:", tx.Hash().String())
		txs = append(txs, tx.Hash().String())
	}
	testutils.GenerateBlocksWaitForTxs(t, h, txs)
	for i := 3; i < 5; i++ {
		target := h.GetWalletManager().GetAccountByIdx(i).EvmAcct.Address
		ba, err = tokenCall.BalanceOf(&bind.CallOpts{}, target)
		if err != nil {
			t.Fatal(err)
		}
		log.Println(i, "address", target.String(), "balance", ba)
		assert.Equal(t, ba, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
	}
	// check transferFrom

	// not approve
	authCaller1, err := testutils.AuthTrans(h.GetBuilder().Get(1))
	if err != nil {
		t.Fatal(err)
	}

	tx1, err := tokenCall.TransferFrom(authCaller1, acc.EvmAcct.Address, h.GetWalletManager().GetAccountByIdx(1).EvmAcct.Address, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
	if err == nil {
		testutils.GenerateBlocksWaitForTxs(t, h, []string{tx1.Hash().String()})
		// check the transaction is ok or not
		txD2, err := h.GetEvmClient().TransactionReceipt(context.Background(), tx1.Hash())
		if err == nil && txD2.Status == 0x1 {
			t.Fatal("Token Bug,TransferFrom without approve")
		}
	}

	//  approve
	txAp, err := tokenCall.Approve(authCaller, h.GetWalletManager().GetAccountByIdx(1).EvmAcct.Address, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
	if err != nil {
		t.Fatal("approve error", err)
	}

	testutils.GenerateBlocksWaitForTxs(t, h, []string{txAp.Hash().String()})
	txLast, err := tokenCall.TransferFrom(authCaller1, acc.EvmAcct.Address, h.GetWalletManager().GetAccountByIdx(1).EvmAcct.Address, big.NewInt(toAmount).Mul(big.NewInt(toAmount), big.NewInt(1e18)))
	if err != nil {
		t.Fatal("TransferFrom error", err)
	}
	testutils.GenerateBlocksWaitForTxs(t, h, []string{txLast.Hash().String()})
	txD3, err := h.GetEvmClient().TransactionReceipt(context.Background(), txLast.Hash())
	if err != nil {
		t.Fatal("TransferFrom error", err)
	}
	assert.Equal(t, txD3.Status, uint64(0x1))
}