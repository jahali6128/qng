// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Qitmeer/qng/common/hash"
	"github.com/Qitmeer/qng/core/json"
	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/core/types/pow"
	"github.com/Qitmeer/qng/engine/txscript"
	"github.com/Qitmeer/qng/testutils/simulator"
)

func GenerateBlockMockNode(t *testing.T, h *simulator.MockNode, num uint64) []*hash.Hash {
	result := make([]*hash.Hash, 0)
	if blocks, err := h.GetPrivateMinerAPI().Generate(uint32(num), pow.MEERXKECCAKV1); err != nil {
		t.Errorf("generate block failed : %v", err)
		return nil
	} else {
		for _, s := range blocks {
			b,_ := hash.NewHashFromStr(s)
			result = append(result, b)
			t.Logf("%v: generate block [%v] ok", h.ID(), b)
		}
	}
	return result
}

// GenerateBlock will generate a number of blocks by the input number for
// the appointed test harness.
// It will return the hashes of the generated blocks or an error
func GenerateBlock(t *testing.T, h *Harness, num uint64) []*hash.Hash {
	result := make([]*hash.Hash, 0)
	if blocks, err := h.Client.Generate(num); err != nil {
		t.Errorf("generate block failed : %v", err)
		return nil
	} else {
		for _, b := range blocks {
			result = append(result, b)
			t.Logf("%v: generate block [%v] ok", h.Node.Id(), b)
		}
	}
	return result
}

func AssertBlockOrderAndHeightMockNode(t *testing.T, h *simulator.MockNode, order, total, height uint64) {
	// order
	if c, err := h.GetPublicBlockAPI().GetBlockCount(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := order
		if c.(uint) != uint(expect) {
			t.Errorf("test failed, expect %v , but got %v", expect, c)
		}
	}
	// total block
	if tal, err := h.GetPublicBlockAPI().GetBlockTotal(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := total
		if tal.(uint) != uint(expect) {
			t.Errorf("test failed, expect %v , but got %v", expect, tal)
		}
	}
	// main height
	if h, err := h.GetPublicBlockAPI().GetMainChainHeight(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := height
		heightInt,err := strconv.Atoi(h.(string))
		if err != nil{
			t.Errorf("test failed : %v", err)
		}
		if uint(heightInt) != uint(expect) {
			t.Errorf("test failed, expect %v , but got %v", expect, h)
		}
	}
}


// AssertBlockOrderAndHeight will verify the current block order, total block number
// and current main-chain height of the appointed test harness and assert it ok or
// cause the test failed.
func AssertBlockOrderAndHeight(t *testing.T, h *Harness, order, total, height uint64) {
	// order
	if c, err := h.Client.BlockCount(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := order
		if c != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, c)
		}
	}
	// total block
	if tal, err := h.Client.BlockTotal(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := total
		if tal != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, tal)
		}
	}
	// main height
	if h, err := h.Client.MainHeight(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := height
		if h != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, h)
		}
	}
}

// Spend amount from the wallet of the test harness and return tx hash
func Spend(t *testing.T, h *Harness, amt types.Amount, preOutpoint *types.TxOutPoint, lockTime *int64) (*hash.Hash, types.Address) {
	addr, err := h.Wallet.NewAddress()
	if err != nil {
		t.Fatalf("failed to generate new address for test wallet: %v", err)
	}
	t.Logf("test wallet generated new address %v ok", addr.Encode())
	addrScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("failed to generated addr script: %v", err)
	}
	output := types.NewTxOutput(amt, addrScript)

	feeRate := types.Amount{Value: 10, Id: amt.Id}
	txId, err := h.Wallet.PayAndSend([]*types.TxOutput{output}, feeRate, preOutpoint, lockTime)
	if err != nil {
		t.Fatalf("failed to pay the output: %v", err)
	}
	return txId, addr
}

// Spend amount from the wallet of the test harness and return tx hash
func SendSelfMockNode(t *testing.T, h *simulator.MockNode, amt types.Amount, lockTime *int64) *hash.Hash {
	acc  := h.GetWalletManager().GetAccountByIdx(0)
	if acc == nil{
		t.Fatalf("failed to get addr")
		return nil
	}
	txId,err := h.GetWalletManager().SendTx(acc.PKHAddress().String(),json.AddressAmountV3{
		acc.PKHAddress().String():json.AmountV3{
			CoinId: uint16(amt.Id),
			Amount: amt.Value,
		},
	},0,*lockTime)
	if err != nil {
		t.Fatalf("failed to pay the output: %v", err)
	}
	ret,err := hash.NewHashFromStr(txId)
	if err != nil {
		t.Fatalf("failed to get the txid: %v, err:%v", txId,err)
	}
	return ret
}


// Spend amount from the wallet of the test harness and return tx hash
func SendExportTxMockNode(t *testing.T, h *simulator.MockNode,txid string,idx uint32, value int64) *hash.Hash {
	acc  := h.GetWalletManager().GetAccountByIdx(0)
	if acc == nil{
		t.Fatalf("failed to get addr")
		return nil
	}
	rawStr,err := h.GetPublicTxAPI().CreateExportRawTransaction(txid,idx,acc.PKAddress().String(),value)
	if err != nil {
		t.Fatalf("failed to pay the output: %v", err)
	}
	
	signRaw, err := h.GetPrivateTxAPI().TxSign(h.GetBuilder().GetHex(0), rawStr.(string), nil)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
		return nil
	}
	allHighFee := true
	tx, err := h.GetPublicTxAPI().SendRawTransaction(signRaw.(string), &allHighFee)
	if err != nil {
		t.Fatalf("failed to send raw tx: %v", err)
		return nil
	}
	ret,err := hash.NewHashFromStr(tx.(string))
	if err != nil {
		t.Fatalf("failed to decode txid: %v", err)
		return nil
	}
	return ret
}


// Spend amount from the wallet of the test harness and return tx hash
func SendSelf(t *testing.T, h *Harness, amt types.Amount, preOutpoint *types.TxOutPoint, lockTime *int64) *hash.Hash {
	addrScript, err := txscript.PayToAddrScript(h.Wallet.addrs[0])
	if err != nil {
		t.Fatalf("failed to generated addr script: %v", err)
	}
	output := types.NewTxOutput(amt, addrScript)

	feeRate := types.Amount{Value: 10, Id: amt.Id}
	txId, err := h.Wallet.PayAndSend([]*types.TxOutput{output}, feeRate, preOutpoint, lockTime)
	if err != nil {
		t.Fatalf("failed to pay the output: %v", err)
	}
	return txId
}

// Spend amount from the wallet of the test harness and return tx hash
func CanNotSpend(t *testing.T, h *Harness, amt types.Amount, preOutpoint *types.TxOutPoint, lockTime *int64) (*hash.Hash, types.Address) {
	addr, err := h.Wallet.NewAddress()
	if err != nil {
		t.Fatalf("failed to generate new address for test wallet: %v", err)
	}
	t.Logf("test wallet generated new address %v ok", addr.Encode())
	addrScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("failed to generated addr script: %v", err)
	}
	output := types.NewTxOutput(amt, addrScript)

	feeRate := types.Amount{Value: 10, Id: amt.Id}
	_, err = h.Wallet.PayAndSend([]*types.TxOutput{output}, feeRate, preOutpoint, lockTime)
	if err != nil {
		t.Fatalf("lock script error:%v", err)
	}
	return nil, addr
}

// TODO, order and height not work for the SerializedBlock
func AssertTxMinedUseSerializedBlock(t *testing.T, h *Harness, txId *hash.Hash, blockHash *hash.Hash) {
	block, err := h.Client.GetSerializedBlock(blockHash)
	if err != nil {
		t.Fatalf("failed to find block by hash %x : %v", blockHash, err)
	}
	numBlockTxns := len(block.Transactions())
	if numBlockTxns < 2 {
		t.Fatalf("the tx has not been mined, the block should at least 2 tx, but got %v", numBlockTxns)
	}
	minedTx := block.Transactions()[1]
	txHash := minedTx.Tx.TxHash()
	if txHash != *txId {
		t.Fatalf("txId %v not match vs block.tx[1] %v", txId, txHash)
	}
	t.Logf("txId %v minted in block, hash=%v", txId, blockHash)
}

func AssertTxMinedUseNotifierAPI(t *testing.T, h *Harness, txId *hash.Hash, blockHash *hash.Hash) {
	block, err := h.Notifier.GetBlockV2FullTx(blockHash.String(), true)
	if err != nil {
		t.Fatalf("failed to find block by hash %x : %v", blockHash, err)
	}
	numBlockTxns := len(block.Tx)
	if numBlockTxns < 2 {
		t.Fatalf("the tx has not been mined, the block should at least 2 tx, but got %v", numBlockTxns)
	}
	minedTx := block.Tx[1]
	txHash := minedTx.Txid
	if txHash != txId.String() {
		t.Fatalf("txId %v not match vs block.tx[1] %v", txId, txHash)
	}
	t.Logf("txId %v minted in block, hash=%v, order=%v, height=%v", txId, blockHash, block.Order, block.Height)
}

func AssertScan(t *testing.T, h *Harness, maxOrder, scanCount uint64) {
	if h.Wallet.maxRescanOrder != maxOrder {
		t.Fatalf("max scan order %d not match %d", h.Wallet.maxRescanOrder, maxOrder)
	}
	if h.Wallet.ScanCount != scanCount {
		t.Fatalf("scan count %d not match %d", h.Wallet.ScanCount, scanCount)
	}
}

func AssertMempoolTxNotify(t *testing.T, h *Harness, txid, addr string, timeout int) {
	TimeoutFunc(t, func() bool {
		h.Wallet.Lock()
		defer h.Wallet.Unlock()
		if h.Wallet.mempoolTx != nil {
			if _, ok := h.Wallet.mempoolTx[txid]; ok {
				return true
			}
		}
		return false
	}, timeout)

	h.Wallet.Lock()
	defer h.Wallet.Unlock()
	if h.Wallet.mempoolTx == nil {
		t.Fatalf("not match mempool tx")
	}
	if _, ok := h.Wallet.mempoolTx[txid]; !ok {
		t.Fatalf("not has mempool tx %s", txid)
	}
	if h.Wallet.mempoolTx[txid] != addr {
		t.Fatalf("mempool tx vout address %s not match %s", h.Wallet.mempoolTx[txid], addr)
	}
}

func AssertTxConfirm(t *testing.T, h *Harness, txid string, confirms uint64) {
	h.Wallet.Lock()
	defer h.Wallet.Unlock()

	if h.Wallet.confirmTxs == nil {
		t.Fatalf("not have any confirms txs")
	}
	if _, ok := h.Wallet.confirmTxs[txid]; !ok {
		t.Fatalf("not has tx %s", txid)
	}
	if h.Wallet.confirmTxs[txid] < confirms {
		t.Fatalf("tx %s confirms %d not right,should more than %d", txid, h.Wallet.confirmTxs[txid], confirms)
	}
}

func AssertTxNotConfirm(t *testing.T, h *Harness, txid string) {
	h.Wallet.Lock()
	defer h.Wallet.Unlock()

	if h.Wallet.confirmTxs == nil {
		return
	}
	if _, ok := h.Wallet.confirmTxs[txid]; ok {
		t.Fatalf("remove has tx %s failed", txid)
	}
}

func TimeoutFunc(t *testing.T, f func() bool, timeout int) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		t1 := time.NewTicker(time.Duration(timeout) * time.Second)
		defer t1.Stop()
		for {
			select {
			case <-t1.C:
				return
			default:
				if f() {
					return
				}
			}
		}
	}()
	wg.Wait()
	t.Logf("time use:%.4f ms", float64(time.Since(start).Milliseconds()))
}

func ConvertEthToMeer(amount *big.Int) *big.Int {
	// wei 1e18  meer 1e8
	amount = amount.Div(amount, big.NewInt(1e10))
	return amount
}

func GenerateBlockWaitForTxs(t *testing.T, h *Harness, txs []string) error {
	tryMax := 20
	txsM := map[string]bool{}
	for _, tx := range txs {
		txsM[tx] = false
	}
	for i := 0; i < tryMax; i++ {
		ret := GenerateBlock(t, h, 1)
		if len(ret) != 1 {
			t.Fatal("No block")
		}
		if len(txsM) <= 0 {
			return nil
		}
		sb, err := h.Client.GetSerializedBlock(ret[0])
		if err != nil {
			t.Fatal(err)
		}
		for _, tx := range sb.Transactions() {
			_, ok := txsM[tx.Hash().String()]
			if ok {
				txsM[tx.Hash().String()] = true
			}
			if types.IsCrossChainVMTx(tx.Tx) {
				txHash := "0x" + tx.Tx.TxIn[0].PreviousOut.Hash.String()
				_, ok := txsM[txHash]
				if ok {
					txsM[txHash] = true
				}
			}
		}
		all := true
		for _, v := range txsM {
			if !v {
				all = false
			}
		}
		if all {
			return nil
		}
		if i >= tryMax-1 {
			t.Fatal("No block")
		}
	}
	return fmt.Errorf("No block")
}
