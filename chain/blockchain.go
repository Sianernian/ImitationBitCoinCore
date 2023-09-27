package chain

import (
	"PublicChain/transaction"
	"PublicChain/utils"
	"PublicChain/utxoset"
	"PublicChain/wallet"
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"math/big"
)

const BUCKERNAME = "blocks"
const LASTHASH = "lasthash"

/**
 * 定义区块链这个结构体，用于存储产生的区块（内存中)
 */
type BlockChain struct {
	//	Blocks []Block
	//文件操作对象
	DB                 *bolt.DB
	LastBlock          Block           // 最新区块
	IteratorBloockHash [32]byte        //迭代到的区块
	Wallet             *wallet.Wallet  // 钱包
	UTXOSet            utxoset.UTXOSet // utxo管理即操作
}

func NewBlockChain(db *bolt.DB) (BlockChain, error) {
	//为lastblock赋值
	var lastBlock Block
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			bucket, _ = tx.CreateBucket([]byte(BUCKERNAME))
		}
		lastHash := bucket.Get([]byte(LASTHASH))
		if len(lastHash) == 0 {
			return nil
		}
		lastBlockBytes := bucket.Get(lastHash)
		lastBlock, _ = UnSerialize(lastBlockBytes)

		return nil
	})
	blockChain := BlockChain{
		DB:                 db,
		LastBlock:          lastBlock,
		IteratorBloockHash: lastBlock.Hash,
	}

	wlt, err := wallet.LoadWalletFromDB(db)
	if err != nil {
		return blockChain, err
	}
	//把构建的wallet对象赋值个 blockChain大的wallet属性
	blockChain.Wallet = &wlt

	set := utxoset.LoadUTXOSetFromDB(db)
	blockChain.UTXOSet = set

	return blockChain, err
}

/**
 * 创建一个区块链实例，该实例携带一个创世区块
 */
func (chain *BlockChain) CreateChainWithGenesis(coinbase []transaction.Transaction) {
	//先看chain.LastBlock是否为空
	hashBig := new(big.Int)
	hashBig.SetBytes(chain.LastBlock.Hash[:])
	if hashBig.Cmp(big.NewInt(0)) == 1 {
		return
	}
	db := chain.DB

	db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			bucket, _ = tx.CreateBucket([]byte(BUCKERNAME))
		}
		if bucket != nil {
			lastHash := bucket.Get([]byte(LASTHASH))
			if len(lastHash) == 0 {

				genesis := CreateGenesisBlock(coinbase)
				genesisBytes, _ := genesis.Serialize()
				//存创世区块
				bucket.Put(genesis.Hash[:], genesisBytes)
				// 更新最新区块
				bucket.Put([]byte(LASTHASH), genesis.Hash[:])
				chain.LastBlock = genesis
				chain.IteratorBloockHash = genesis.Hash
			}
		}
		return nil
	})

}

func (chain *BlockChain) CreateCoinbase(addr string) ([]byte, error) {
	//1.判断地址有效性
	isValid := wallet.IsAddressValid(addr)
	if !isValid {
		return nil, errors.New("输入地址有误，请重新参试")
	}

	coinbase, err := transaction.NewCoinbaseTx(addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	chain.CreateChainWithGenesis([]transaction.Transaction{*coinbase})

	// 把coinbase交易产生的utxo存到utxoset中
	utxos := make([]transaction.UTXO, 0)
	for index, output := range coinbase.Outputs {
		utxo := transaction.NewUTXO(coinbase.TxHash, index, output)
		utxos = append(utxos, utxo)
	}
	success := chain.UTXOSet.AddUTXOWithAddress(addr, utxos)
	if !success {
		return nil, errors.New("保存utxo添加失败")
	}
	//把address设置为默认矿工地址
	return coinbase.TxHash[:], chain.SetCoinbase(addr)
}

/*
*

	获取某个地址的余额
*/
func (chain *BlockChain) GetBalance(address string) float64 {
	_, totalBalance := chain.GetUtxoWithBalance(address, []transaction.Transaction{})
	return totalBalance
}

/*
*

	获取某个特定的地址余额和 所能花费的utxoSet
*/
func (chain *BlockChain) GetUtxoWithBalance(address string, txs []transaction.Transaction) ([]transaction.UTXO, float64) {
	//文件中遍历区块，找出区块已经存在交易中可花费utxo
	//dbUtxos:=chain.SearchUTXO(address)
	dbUtxos, err := chain.UTXOSet.QuerryUTXOByAddress(address)
	if err != nil {
		fmt.Println(err.Error(), "从db中查询地址可用的utxo失败")
		return nil, -1
	}
	//遍历内存中的 txs切片,如果当前已经构建还未存储的交易花了钱，要删掉
	memSpends := make([]transaction.TxInput, 0)
	memearns := make([]transaction.UTXO, 0)

	kerPair := chain.Wallet.GetKeyPairByAddress(address)
	if kerPair == nil {
		return nil, 0
	}
	for _, tx := range txs {

		for _, input := range tx.Inputs {
			if input.CheckPubKeyWithpubKey(kerPair.Pub) {
				memSpends = append(memSpends, input)
			}
		}
		for index, output := range tx.Outputs {
			if output.VertifyOutputWithAddress(address) {
				utxo := transaction.NewUTXO(tx.TxHash, index, output)
				memearns = append(memearns, utxo)
			}
		}
	}

	// 将内存中以花的utxo从dbutxo删掉，将内存中产生的收入加入到可花费收入中
	utxos := make([]transaction.UTXO, 0)
	var isSpent bool
	for _, dbUtxo := range dbUtxos {
		isSpent = false
		for _, memUtxo := range memSpends {
			//判断某个UTXO 是否已经被消费掉

			if dbUtxo.IsSpent(memUtxo) {
				isSpent = true
			}
		}
		if !isSpent {
			utxos = append(utxos, dbUtxo)
		}
	}

	utxos = append(utxos, memearns...)

	fmt.Printf("地址%s一共找到%d笔金额\n", address, len(utxos))
	var totalBalance float64
	for _, utxo := range utxos {
		totalBalance += utxo.Value
	}
	fmt.Printf("获取%s的utxo和余额，余额是：%f\n", address, totalBalance)
	return utxos, totalBalance
}

func (chain *BlockChain) SendTransaction(from string, to string, value string) error {
	fromSlice, err := utils.JsonStringToSlince(from)
	toSlice, err := utils.JsonStringToSlince(to)
	valueSlice, err := utils.JsonFloatToSlice(value)
	if err != nil {
		return err
	}

	//判断参数的长度，筛选参数不匹配的情况
	lenFrom := len(fromSlice)
	lenTo := len(toSlice)
	lenValue := len(valueSlice)
	if !(lenFrom == lenTo && lenFrom == lenValue) {
		return errors.New("发起交易的参数不匹配，请检查后重试")
	}

	//地址有效性的判断
	for i := 0; i < len(fromSlice); i++ {
		//交易发起人的地址是否合法，合法为true，不合法为false
		isFromValid := wallet.IsAddressValid(fromSlice[i])
		//交易接收者的地址是否合法，合法为true，不合法为false
		isToValid := wallet.IsAddressValid(toSlice[i])
		//from: 合法   合法
		//to:   不合法  不合法
		if !isFromValid || !isToValid {
			return errors.New("交易的参数地址不合法，请检查后重试")
		}
	}

	//遍历参数的切片，创建交易
	txs := make([]transaction.Transaction, 0)
	for index := 0; index < lenFrom; index++ {
		utxos, totalBalance := chain.GetUtxoWithBalance(fromSlice[index], txs)
		//fmt.Printf("转账发起人%s,当前余额：%f,接收者:%s,转账数额：%f\n", fromSlice[index], totalBalance, toSlice[index], valueSlice[index])
		if totalBalance < valueSlice[index] {
			return errors.New("抱歉，" + fromSlice[index] + "余额不足，请充值！")
		}

		var inputAmount float64 //总的花费的钱数

		utxoNum := 0
		for num, utxo := range utxos {
			inputAmount += utxo.Value
			if inputAmount >= valueSlice[index] {
				//够花了
				utxoNum = num
				break
			}
		}
		keyPair := chain.Wallet.GetKeyPairByAddress(fromSlice[index])
		//1、创建交易

		tx, err := transaction.NewTransaction(
			utxos[:utxoNum+1],
			fromSlice[index],
			keyPair.Pub,
			toSlice[index],
			valueSlice[index])
		if err != nil {
			return errors.New("抱歉，创建交易失败，请检查后重试")
		}

		//2、使用from对应的私钥对tx进行交易签名
		err = tx.Sign(keyPair.Pri, utxos[:utxoNum+1])
		//如果任何一笔交易签名失败，则全部交易结束，返回错误信息
		if err != nil {
			return err
		}
		//3、把已经构建好并且签好名的交易放入到txs内存切片容器中
		txs = append(txs, *tx)
	}

	address := chain.GetCoinbase()
	if len(address) == 0 {
		return errors.New("未设置coinbase矿工地址，请先设置")
	}
	coninbase, err := transaction.NewCoinbaseTx(address)
	if err != nil {
		return nil
	}

	sumTxs := make([]transaction.Transaction, 0)
	sumTxs = append(sumTxs, *coninbase)
	sumTxs = append(sumTxs, txs...)
	//对即将要打包到区块中的交易进行签名验证，确保交易的正确性。
	//如果发现有非法的交易（即签名验证失败），则终止打包，返回错误。

	//该段验证签名的代码由矿工节点执行，对每一笔交易依次进行签名
	for _, tx := range txs {

		//首先判断交易是否是coinbase交易，如果是，则不需要验签
		if tx.IsCoinbaseTranaction() {
			continue
		}
		//1、先找出当前的交易tx消费的是哪些utxo
		spendUTXOs, err := chain.FindSpentUTXOsByTrabsaction(tx, txs)
		if err != nil {
			return err
		}

		//2、把找到的该笔tx所消费的utxo传入到签名验证方法中，供进行使用
		verify, err := tx.VertifySign(spendUTXOs)
		if err != nil {
			return err
		}

		if !verify {
			return errors.New("交易失败。请重试！")
		}
	}

	//把构建好的交易存入到区块中
	err = chain.AddNewBlock(txs)
	if err != nil {
		return err
	}
	// txs: coinbase + 用户自定义交易

	//遍历txs，统计哪些地址，产生了哪些utxo 将结果保存
	usxoSet := make(map[string][]transaction.UTXO)
	for txindex, tx := range txs {
		for index, output := range tx.Outputs {
			utxo := transaction.NewUTXO(tx.TxHash, index, output)
			isSpent := false
			for i := txindex + 1; i < len(txs); i++ {
				for _, input := range txs[i].Inputs {
					if utxo.IsSpent(input) {
						isSpent = true
					}
				}
			}
			if !isSpent {
				//根据utxo.pubhash得到地址
				address := chain.Wallet.GetAddressByPubKHash(utxo.PubHash)

				utxos := usxoSet[address]
				if len(utxos) == 0 {
					utxos = make([]transaction.UTXO, 0)
				}
				utxos = append(utxos, utxo)
				usxoSet[address] = utxos
			}
		}
	}
	for pubkHash, utxos := range usxoSet {
		success := chain.UTXOSet.AddUTXOWithAddress(pubkHash, utxos)
		if !success {
			return errors.New("保存失败")
		}
	}

	//从UTXOSet中把已消费掉的utxo删除掉
	// target ：统计当前内存中交易哪些地址花了哪些utxo，统计即可
	spendRecords := make(map[string][]utxoset.SpendRecord, 0)

	for _, tx := range txs {
		for _, input := range tx.Inputs {
			//只需要记录每个input消费的txid 和 vout
			spendRecord := utxoset.NewSpendRecord(input.Txid, input.Vout)
			address, err := wallet.NewAddress(input.Pubk)
			if err != nil {
				return err
			}
			records := spendRecords[address]
			if len(records) == 0 {
				records = make([]utxoset.SpendRecord, 0)
			}
			records = append(records, spendRecord)

			spendRecords[address] = records
		}
	}

	// 对每个地址消费统计的spendrecord进行删除
	for address, record := range spendRecords {
		success := chain.UTXOSet.Change(address, record)
		if !success {
			return errors.New("更新utxo数据失败")
		}
	}

	return nil
}

func (chain *BlockChain) AddNewBlock(txs []transaction.Transaction) error {
	//1.从db中找到最后一个区块数据
	db := chain.DB
	//2. 获取到最新区块
	lastBlock := chain.LastBlock

	//3. 得到区块属性
	newBlock, err := CreateBlock(lastBlock.Height, lastBlock.Hash, txs)
	if err != nil {
		return err
	}
	newBlockBytes, err := newBlock.Serialize()
	if err != nil {
		return err
	}
	//4. 更新db文件，将新生成的区块写入文件中
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			err = errors.New("上一步有错")
			return err
		}
		//更新区块数据
		bucket.Put(newBlock.Hash[:], newBlockBytes)
		//更新最新区块指向标记
		bucket.Put([]byte(LASTHASH), newBlock.Hash[:])
		return nil

		//更新blockChain对象的lastBlock结构体
		chain.LastBlock = *newBlock
		chain.IteratorBloockHash = lastBlock.Hash

		return nil
	})
	return nil

}

func (chain BlockChain) GetLastBlock() Block {
	return chain.LastBlock
}

/*
*

	获取所有区块
*/
func (chain BlockChain) GetAllBlocks() ([]Block, error) {
	db := chain.DB
	blocks := make([]Block, 0)
	var err error
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			err = errors.New("区块数据区操作失败")
			return err
		}
		blocks = append(blocks, chain.LastBlock)
		//lastHash:=bucket.Get([]byte(LASTHASH))
		var currentHash []byte
		hash := [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		//直接从倒数第二个区块进行遍历
		currentHash = chain.LastBlock.PreHash[:]
		for {
			//倒数第二个区块开始遍历
			currentBlockBytes := bucket.Get(currentHash)
			currentBlock, _ := UnSerialize(currentBlockBytes)

			blocks = append(blocks, currentBlock)

			currentHash = currentBlock.PreHash[:]
			if bytes.Compare(hash[:], currentHash) == 0 {

				break
			}
		}
		return nil
	})
	return blocks, err
}

/*
*

	该方法用于实现Iterator迭代器接口的方法  判断是否还有区块
*/
func (chain *BlockChain) IsNext() bool {
	//是否还有前一个区块
	engine := chain.DB
	var isNext bool
	genesishash := [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	engine.View(func(tx *bolt.Tx) error {
		currentBlockHash := chain.IteratorBloockHash[:]

		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			return errors.New("区块数据文件失败")
		}
		currentBlockByte := bucket.Get(currentBlockHash[:])
		currentBlock, _ := UnSerialize(currentBlockByte)
		if bytes.Compare(currentBlock.Hash[:], genesishash[:]) == 0 {
			isNext = false
		} else {
			isNext = true
		}

		//isNext = len(preBlockBytes) !=0
		return nil
	})
	return isNext
}

/*
*

	该方法用于实现Iterator迭代器接口的方法，用于取出下一个区块（前一个区块）
*/
func (chain *BlockChain) Next() Block {
	engine := chain.DB
	var currentBlock Block
	engine.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(BUCKERNAME))
		if bucket == nil {
			return errors.New("区块数据文件失败")
		}
		currentBlockBytes := bucket.Get(chain.IteratorBloockHash[:])
		currentBlock, _ = UnSerialize(currentBlockBytes)
		chain.IteratorBloockHash = currentBlock.PreHash
		return nil
	})
	return currentBlock
}

/**
  定义该方法，用于实现寻找与from有关的所有可花费的交易输出，既UTXO
*/

func (chain BlockChain) SearchUTXO(from string) []transaction.UTXO {

	//定义容器，存放from所有的花费
	spents := make([]transaction.TxInput, 0)
	//定义容器，存放from所有的收入
	earns := make([]transaction.UTXO, 0)

	keyPair := chain.Wallet.GetKeyPairByAddress(from)
	if keyPair == nil {
		return nil
	}
	for chain.IsNext() { //遍历区块
		block := chain.Next()
		for _, tx := range block.Txs { //遍历交易
			//a.遍历交易输入
			for _, input := range tx.Inputs {
				if input.CheckPubKeyWithpubKey(keyPair.Pub) {
					continue
				} //from花费了
				spents = append(spents, input)
			}
			//b.遍历交易输出
			for index, output := range tx.Outputs {
				if !output.VertifyOutputWithAddress(from) {
					continue
				}
				//交易输出是from的，有收入
				input := transaction.NewUTXO(tx.TxHash, index, output)

				earns = append(earns, input)
			}
		}
	}
	fmt.Printf("地址%s的所有收入有%d笔\n", from, len(earns))
	fmt.Printf("地址%s的所有花费有%d笔\n", from, len(spents))

	utxos := make([]transaction.UTXO, 0)
	var earnOrSpent bool
	// 遍历spent 和 earn ，将已花费的记录剔除，剩下UTXO
	for _, earn := range earns {
		earnOrSpent = false
		//判断每一笔交易是否在之前交易中已被花费
		for _, spent := range spents {
			if earn.TxId == spent.Txid && earn.Vout == spent.Vout {
				earnOrSpent = true
				break
			}
		}
		if !earnOrSpent {
			utxos = append(utxos, earn)
		}
	}
	return utxos
}

func (chain *BlockChain) GetNewAddress() (string, error) {
	fmt.Println("chain.DB:", chain.DB)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()
	chain.Wallet.DB = chain.DB //panic: runtime error: invalid memory address or nil pointer dereference
	return chain.Wallet.CreateNewAddress()
}

func (chain *BlockChain) GetAddressList() ([]string, error) {
	if chain.Wallet == nil {
		return nil, errors.New("钱包报错，请重试")
	}
	addlist := make([]string, 0)
	for address, _ := range chain.Wallet.Address {
		addlist = append(addlist, address)
	}
	return addlist, nil
}

func (chain *BlockChain) DumpPrivateKey(add string) (*ecdsa.PrivateKey, error) {
	isAddressValid := wallet.IsAddressValid(add)
	if !isAddressValid {
		return nil, errors.New("地址不合法，请重试")
	}

	if chain.Wallet == nil {
		return nil, errors.New("钱包报错，请重试")
	}
	keyPair := chain.Wallet.Address[add]
	if keyPair == nil {
		return nil, errors.New("未找到对应的地址私钥")
	}
	return keyPair.Pri, nil
}

/**
 * 该方法用于接收特定的交易tx，用于寻找到某个交易具体花费了哪些utxo，并寻找到结构
 */
func (chain BlockChain) FindSpentUTXOsByTrabsaction(transac transaction.Transaction, memTxs []transaction.Transaction) ([]transaction.UTXO, error) {
	//消费utxo可能来源途径 1。持久化存储数据区块中 1.1 更换 到 utxoSet中寻找  2. 内存中
	var err error
	spentUTXOs := make([]transaction.UTXO, 0)
	// 先去持久化的区块当中去找消费的utxo
	//for chain.IsNext() {
	//	block := chain.Next()
	//	chain.DB.View(func(tx *bolt.Tx) error {
	//		bucket := tx.Bucket([]byte(BUCKERNAME))
	//		if bucket == nil {
	//			err = errors.New("查询交易UTXO失败")
	//			return err
	//		}
	//		//遍历区块中的交易
	//		for _, tran := range block.Txs {
	//
	//			//遍历output,因为output代表的是可花费的钱，即有可能在transac中被input所引用
	//			for outIndex, output := range tran.Outputs {
	//				utxo := transaction.NewUTXO(tran.TxHash, outIndex, output)
	//
	//				//每次找到一个output即交易输出，都应该到transac的交易输入中去核查一遍，看看是否被消费了
	//				for _, input := range transac.Inputs {// 遍历db交易
	//					if utxo.IsSpent(input) {
	//						spentUTXOs = append(spentUTXOs, utxo)
	//					}
	//				}
	//			}
	//		}
	//		return nil
	//	})
	//}

	// 去utxo Set中寻找 该交易花费的utxo
	records := make([]utxoset.SpendRecord, 0)

	for _, input := range transac.Inputs {
		spendRecord := utxoset.NewSpendRecord(input.Txid, input.Vout)
		records = append(records, spendRecord)
	}
	// 不是coinbase  transac.Inputs[0]不会报错 可以不用判断input长度
	address, err := wallet.NewAddress(transac.Inputs[0].Pubk)
	if err != nil {
		return nil, err
	}
	spentUTXOs, err = chain.UTXOSet.QuerrySpendUTXOs(address, records)
	if err != nil {
		return nil, err
	}

	//内存中
	for _, memTx := range memTxs {
		// 只遍历交易本身以外的内存中的其他地址
		if bytes.Compare(memTx.TxHash[:], transac.TxHash[:]) == 0 {
			continue
		}
		for outIndex, output := range memTx.Outputs {
			utxo := transaction.NewUTXO(memTx.TxHash, outIndex, output)
			for _, input := range transac.Inputs {

				if utxo.IsSpent(input) {
					spentUTXOs = append(spentUTXOs, utxo)
				}
			}
		}
	}

	//最终把，①和② 两个渠道找到的当前该表交易所花费的utxo，进行返回
	return spentUTXOs, nil
}

// 该方法用于设置用户自定义的矿工地址
func (chain *BlockChain) SetCoinbase(address string) error {
	//1.现做地址的规范性校验
	if !wallet.IsAddressValid(address) {
		return errors.New("地址有误，请输入正确的地址")
	}

	return chain.Wallet.SetCoinbase(address)
}

func (chain *BlockChain) GetCoinbase() string {
	return chain.Wallet.GetCoinbase()
}
