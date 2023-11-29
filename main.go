package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sync/atomic"

	//"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"crypto/rand"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

var (
	priv      *ecdsa.PrivateKey
	address   common.Address
	ethClient *ethclient.Client
	dataTemp  string
)
var (
	globalNonce = time.Now().UnixNano()
	zeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
	chainID     = big.NewInt(0)
	userNonce   = -1
)

func main() {
	log.Infoln()
	log.Info(`
Author:[𝕏] @chenmin22998595
Author:[𝕏] @chenmin22998595
Author:[𝕏] @chenmin22998595
`)
	log.Infoln()
	dataTemp = fmt.Sprintf(`data:application/json,{"p":"rerc-20","op":"mint","tick":"%s","id":"%%s","amt":"%d"}`, config.Tick, config.Amt)
	var err error
	ethClient, err = ethclient.Dial(config.Rpc)
	if err != nil {
		panic(err)
	}

	chainID, err = ethClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	bytePriv, err := hexutil.Decode(config.PrivateKey)
	if err != nil {
		panic(err)
	}
	prv, _ := btcec.PrivKeyFromBytes(bytePriv)
	priv = prv.ToECDSA()
	address = crypto.PubkeyToAddress(*prv.PubKey().ToECDSA())
	log.WithFields(log.Fields{
		"prefix":   config.Prefix,
		"amt":      config.Amt,
		"tick":     config.Tick,
		"count":    config.Count,
		"address":  address.String(),
		"chain_id": chainID.Int64(),
	}).Info("prepare done")

	startNonce := globalNonce
	go func() {
		for {
			last := globalNonce
			time.Sleep(time.Second * 10)
			log.WithFields(log.Fields{
				"hash_rate":  fmt.Sprintf("%dhashes/s", (globalNonce-last)/10),
				"hash_count": globalNonce - startNonce,
			}).Info()
		}
	}()

	wg := new(sync.WaitGroup)
	for i := 0; i < config.Count; i++ {
		tx := makeBaseTx()
		//log.Info(tx)
		wg.Add(runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		for j := 0; j < runtime.NumCPU(); j++ {
			go func(ctx context.Context, cancelFunc context.CancelFunc) {
				for {
					select {
					case <-ctx.Done():
						wg.Done()
						return
					default:
						makeTx(cancelFunc, tx)
					}
				}
			}(ctx, cancel)
		}
		wg.Wait()
	}
}

func makeTx(cancelFunc context.CancelFunc, innerTx *types.DynamicFeeTx) {
	atomic.AddInt64(&globalNonce, 1)

	potential_solution, _ := generateRandomHash()
	hash := potential_solution + "7245544800000000000000000000000000000000000000000000000000000000"
	data, _ := decodeHex(hash)
	hashBytes := keccak256Hash(data)
	hashed_solution := hex.EncodeToString(hashBytes)
	//fmt.Println(hash, hashString)

	if strings.HasPrefix(hashed_solution, config.Prefix) {
		log.WithFields(log.Fields{
			"hashed_solution":    hashed_solution,
			"potential_solution": potential_solution,
		}).Info("new hash")

		temp := fmt.Sprintf(dataTemp, potential_solution)
		innerTx.Data = []byte(temp)
		tx := types.NewTx(innerTx)
		signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainID), priv)

		err := ethClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.WithFields(log.Fields{
				"tx_hash": signedTx.Hash().String(),
				"err":     err,
			}).Error("failed to send transaction")
		} else {
			log.WithFields(log.Fields{
				"tx_hash": signedTx.Hash().String(),
			}).Info("broadcast transaction")
		}

		cancelFunc()
	}
}

func makeBaseTx() *types.DynamicFeeTx {
	if userNonce < 0 {
		nonce, err := ethClient.PendingNonceAt(context.Background(), address)
		if err != nil {
			panic(err)
		}
		userNonce = int(nonce)
	} else {
		userNonce++
	}
	innerTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     uint64(userNonce),
		GasTipCap: new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(int64(config.GasTip))),
		GasFeeCap: new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(int64(config.GasMax))),
		Gas:       32000,
		To:        &address,
		Value:     big.NewInt(0),
	}

	return innerTx
}

// decodeHex 解码十六进制字符串
func decodeHex(hexStr string) ([]byte, error) {
	// 移除可能的"0x"前缀
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
}

// keccak256Hash 计算Keccak256哈希值
func keccak256Hash(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// generateRandomHash 生成一个随机的SHA256哈希值，并返回其十六进制字符串表示
func generateRandomHash() (string, error) {
	// 生成随机数据
	randomData := make([]byte, 32)
	_, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}

	// 计算哈希值
	hash := sha256.Sum256(randomData)
	//fmt.Println(hash)
	// 转换哈希值为十六进制字符串
	hexHash := hex.EncodeToString(hash[:])

	return "0x" + hexHash, nil
}
