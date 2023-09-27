package utxoset

// 表示一笔消费记录
type SpendRecord struct {
	TxId  [32]byte
	Vout  int
}


func NewSpendRecord(txid [32]byte, vout int)SpendRecord{
	return SpendRecord{
		TxId: txid,
		Vout: vout,
	}
}

func(spendRecord SpendRecord)GetTxId()[32]byte{
	return spendRecord.TxId
}
func(spendRecord SpendRecord)GetVout()int{
	return spendRecord.Vout
}