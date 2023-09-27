package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"golang.org/x/crypto/ripemd160"
)

/*
  将一个int64数值类型转换【】byte
 */
func IntToByte(num int64)([]byte,error){
	buff :=new(bytes.Buffer)
	err :=binary.Write(buff,binary.BigEndian,num)
	if err !=nil{
		return nil ,err
	}
	return buff.Bytes(),nil
}


/**
   公用的gob序列化
 */
func GobEncode(entity interface{})([]byte ,error){
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)
	err := encoder.Encode(entity)
	return buff.Bytes(), err
}

/**
     公共的gob反序列化
 */
func GodDecode(data []byte,entity interface{})(interface{} ,error){
	decode :=gob.NewDecoder(bytes.NewReader(data))
	err :=decode.Decode(entity)
	return entity,err
	//json.Unmarshal()
}

func JsonStringToSlince(data string)([]string,error){
	var slice []string
	err :=json.Unmarshal([]byte(data),&slice)
	return slice,err
}

//JsonArray [10.0,20.1] ---> []float[10.0  20.0]
func JsonFloatToSlice(data string)([]float64 ,error){
	var slice []float64
	err :=json.Unmarshal([]byte(data),&slice)
	return slice ,err
}

func Sha256Hash(data []byte) []byte{
	sha256Hash :=sha256.New()
	sha256Hash.Write(data)
	return sha256Hash.Sum(nil)
}

func Ripemd160(data []byte)[]byte{
	Ripemd160:=ripemd160.New()
	Ripemd160.Write(data)
	return Ripemd160.Sum(nil)
}



