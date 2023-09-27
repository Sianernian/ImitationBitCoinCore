package utxoset

import (
	"PublicChain/transaction"
	"PublicChain/utils"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
)

const UTXOSET = "utxoset"

/*
定义一个UTXOSet结构体，该结构体用于管理单个的UtXO

	a.将某个地址新增的utxo的数据存储到文件中
	b.从UTXOSet整个集合中删除掉某个地址address消费了的utxo
*/
type UTXOSet struct {
	//UTXO map[string][]transaction.UTXO
	DB *bolt.DB
}

// 用于找出指定地址在某次交易中所消费的utxo
func (utxoset *UTXOSet) QuerrySpendUTXOs(address string, spends []SpendRecord) ([]transaction.UTXO, error) {
	db := utxoset.DB
	var err error
	exisutxos := make([]transaction.UTXO, 0)
	spentUTXOs := make([]transaction.UTXO, 0)
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(UTXOSET))
		if bucket == nil {
			err = errors.New("UtXO查询失败")
			return err
		}
		utxosBytes := bucket.Get([]byte(address))
		if len(utxosBytes) == 0 {
			return nil
		}
		//utxo 不为nil 尝试找符合条件utxo

		decoder := gob.NewDecoder(bytes.NewReader(utxosBytes))
		err = decoder.Decode(&exisutxos)
		if err != nil {
			return err
		}

		isContaines := IsSubsetUtXOs(exisutxos, spends)
		if !isContaines {
			err = errors.New("查找消费的utxo有误")
			return err
		}

		// 遍历 花费掉了的utxo 记录下来
		for _, utxo := range exisutxos {
			for _, record := range spends {
				if utxo.EqualSpendRecord(record) {
					// 满足 条件 表示当前找到的utxo在当钱交易中被消费
					spentUTXOs = append(spentUTXOs, utxo)

				}
			}
		}

		return nil
	})
	return spentUTXOs, err
}

/*
*

	在构建交易之前，先查询到某个地址所拥有的所有可用utxo和总共余额，直接从utxoSet取出
*/
func (utxoset *UTXOSet) QuerryUTXOByAddress(address string) ([]transaction.UTXO, error) {
	db := utxoset.DB
	var err error
	utxos := make([]transaction.UTXO, 0)
	fmt.Println("utxoSet:", db)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(UTXOSET))
		if bucket == nil {
			err = errors.New("UtXO查询失败")
			return err
		}
		utxosBytes := bucket.Get([]byte(address))
		if len(utxosBytes) == 0 {
			return nil
		}
		decoder := gob.NewDecoder(bytes.NewReader(utxosBytes))
		err = decoder.Decode(&utxos)

		return err
	})
	if err != nil {
		return nil, err
	}

	return utxos, err
}

/*
*

	将要交易存到区块中，存入交易中产生新的utxo
*/
func (utxoset *UTXOSet) AddUTXOWithAddress(address string, utxos []transaction.UTXO) bool {
	db := utxoset.DB
	var err error
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(UTXOSET))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(UTXOSET))
			if err != nil {
				return err
			}
		}
		//先取
		exisUTXOs := make([]transaction.UTXO, 0)
		utxosBytes := bucket.Get([]byte(address))
		if len(utxosBytes) != 0 {
			decoder := gob.NewDecoder(bytes.NewReader(utxosBytes))
			err = decoder.Decode(&exisUTXOs)
			if err != nil {
				return err
			}

		}
		// 合并已有和新增
		exisUTXOs = append(exisUTXOs, utxos...)

		utxosBytes, err := utils.GobEncode(&exisUTXOs)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(address), utxosBytes)
		return err
	})
	return err == nil
}

/*
*

	将某个地址对应已经消费了的utxo数据从utxoSet中删除
*/
func (utxoset *UTXOSet) Change(address string, records []SpendRecord) bool {
	db := utxoset.DB
	var err error
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(UTXOSET))
		if bucket == nil {
			return nil
		}

		utxosBytes := bucket.Get([]byte(address))
		if len(utxosBytes) == 0 {
			err = errors.New("删除utxo记录失败")
			return err
		}
		existUTXOs := make([]transaction.UTXO, 0)
		decoder := gob.NewDecoder(bytes.NewReader(utxosBytes))
		err = decoder.Decode(&existUTXOs)
		if err != nil {
			return err
		}

		// 要删除的utxo都存在
		isContains := IsSubsetUtXOs(existUTXOs, records)
		if !isContains {
			err = errors.New("招不到要删除的utxo,请检查")
			return err
		}
		//消费剩下的utxo
		remainUTXOs := make([]transaction.UTXO, 0)
		for _, utxo := range existUTXOs { // 当前地址目前存在的utxo
			isSpent := false
			for _, reccord := range records {
				if utxo.EqualSpendRecord(reccord) {
					isSpent = true
				}
			}
			if !isSpent {
				remainUTXOs = append(remainUTXOs, utxo)
			}
		}
		// 把剩下存储起来
		remainBytes, err := utils.GobEncode(remainUTXOs)
		if err != nil {
			return err
		}
		bucket.Put([]byte(address), remainBytes)
		return nil
	})
	return err == nil
}

/*
*

	判断俩个utxo集合  返回结果为sub是否是all子集
*/
func IsSubsetUtXOs(all []transaction.UTXO, sub []SpendRecord) bool {
	for _, subset_utxo := range sub {
		isCointains := false
		for _, super_utxo := range all {
			if super_utxo.EqualSpendRecord(subset_utxo) {
				isCointains = true
			}
		}
		if !isCointains {
			return false
		}
	}
	return true
}

// 实例化 utxoset
func LoadUTXOSetFromDB(db *bolt.DB) UTXOSet {
	return UTXOSet{db}
}
