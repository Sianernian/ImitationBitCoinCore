package client

import (
	"PublicChain/chain"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"
)

/**
  客户端（命令行窗口工具）， 主要用户实现与用户进行动态交互
       1. 将帮助信息等输出到控制台
       2. 读取参数并解析，根据解析结果调用对应的项目功能
*/

type Client struct {
	Chain chain.BlockChain
}

/*
*

	client的run方法，是程序的主要处理逻辑入口
*/
func (client *Client) Run() {
	if len(os.Args) == 1 {
		client.Help()
		return
	}
	//1.解析命令行参数
	command := os.Args[1]
	//2.确定用户输入命令
	switch command {
	case CREATECHAIN:
		client.CreateChain()
	case GENERATEGENESIS: //发送第一笔交易
		client.GenerateGenesis()
	case SENDTRASACTION: //发送新的交易
		client.SendTransaction()
	case GETLASTBLOCK: // 得到最新区块
		client.GetLastBlock()
	case GETBLOCKCOUNT: // 记录区块多少
		client.GetBlockCount()
	case GETALLBLOCKS: //查询所有区块
		client.GetAllBlocks()
	case HELP: //帮助
		client.Help()
	case GETBALANCE: // 得到余额
		client.GetBalance()
	case GETNEWADDRESS: //生成新的地址
		client.GetNewAddress()
	case LISTADDRESS: // 打印出地址列表
		client.ListAddress()
	case DUMPPRIVATEKEY:
		client.DumpPrivateKey()
	case SETCOINBASE: // 设置coinbase地址
		client.SetCoinBase()
	case GETCOINBASE: //得到coinbase地址
		client.GetCoinBase()
	default:
		client.Default()
	}

}

// 用于接收用户的参数，设置矿工地址
func (client *Client) SetCoinBase() {
	setCoinBase := flag.NewFlagSet(SETCOINBASE, flag.ExitOnError)
	coinbase := setCoinBase.String("address", "", "用户设置矿工地址")
	_ = setCoinBase.Parse(os.Args[2:])
	if len(os.Args[2:]) > 2 {
		fmt.Println("不支持参数")
		return
	}
	_ = client.Chain.SetCoinbase(*coinbase)
}

// 获取当前节点设置的coinbase地址，并返回给用户
func (client *Client) GetCoinBase() {
	getCoinbase := flag.NewFlagSet(GETCOINBASE, flag.ExitOnError)
	_ = getCoinbase.Parse(os.Args[2:])
	if len(os.Args[2:]) > 0 {
		fmt.Println("无效参数,请重新尝试")
		return
	}
	coinbase := client.Chain.GetCoinbase()
	if len(coinbase) == 0 {
		fmt.Println("未获取到coinbase地址，请先设置")
		return
	}
	fmt.Printf("Miner得到的地址是：%s\n", coinbase)

}

func (client *Client) CreateChain() {
	CreateChain := flag.NewFlagSet(CREATECHAIN, flag.ExitOnError)
	fmt.Println("CreateChain :", CreateChain)
	parse := CreateChain.Parse(os.Args[1:])
	fmt.Println("第二个参数parse", parse)

}

func (client *Client) DumpPrivateKey() {
	dumpPrivatrKey := flag.NewFlagSet(DUMPPRIVATEKEY, flag.ExitOnError)
	address := dumpPrivatrKey.String("address", "", "要导出的地址")
	_ = dumpPrivatrKey.Parse(os.Args[2:])

	if len(os.Args[2:]) > 2 {
		fmt.Println("无法解析输入的参数，请重试")

		return
	}
	privateKey, err := client.Chain.DumpPrivateKey(*address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("私钥为:%x\n", privateKey.D.Bytes())

}
func (client *Client) ListAddress() {
	listAddress := flag.NewFlagSet(LISTADDRESS, flag.ExitOnError)
	_ = listAddress.Parse(os.Args[2:])
	if len(os.Args[2:]) > 0 {
		fmt.Println("无法解析输入的参数，请重新尝试")
		return
	}
	addList, err := client.Chain.GetAddressList()
	if err != nil {
		fmt.Println("加载地址失败：", err.Error())
		return
	}
	if len(addList) == 0 {
		fmt.Println("当前暂时无地址")
		return
	}
	fmt.Println("地址列表获取成功，信息如下：")
	for index, address := range addList {
		fmt.Printf("(%d) : %s\n", index+1, address)
	}
}

func (client *Client) GetNewAddress() {
	getNewAddress := flag.NewFlagSet(GETNEWADDRESS, flag.ExitOnError)
	err := getNewAddress.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 判断用户是否输入参数，getnewaddress 不需要参数
	if len(os.Args[2:]) > 0 {
		fmt.Println("getnewaddress 不需要参数，请重试哟！")
		return
	}
	address, _ := client.Chain.GetNewAddress()
	fmt.Println("生成的地址是:", address)
}

func (client *Client) GenerateGenesis() {
	generateGensis := flag.NewFlagSet(GENERATEGENESIS, flag.ExitOnError)
	address := generateGensis.String("address", "", "用户指定地址")
	_ = generateGensis.Parse(os.Args[2:])
	// 先判断是否已存在创世区块
	hashBig := new(big.Int)
	hashBig.SetBytes(client.Chain.LastBlock.Hash[:])
	if hashBig.Cmp(big.NewInt(0)) == 1 {
		fmt.Println("创世区块已存在，无法覆盖")
		return
	}

	//解析
	coinbaseHash, err := client.Chain.CreateCoinbase(*address)
	if err != nil {
		fmt.Println("coinbase交易出现错误，请重试。")
		return
	}

	fmt.Printf("交易hash是:%x\n", coinbaseHash)

}

func (client *Client) GetBalance() {
	var address string
	getbalance := flag.NewFlagSet(GETBALANCE, flag.ExitOnError)
	getbalance.StringVar(&address, "address", "", "要查询地址的余额")
	_ = getbalance.Parse(os.Args[2:])
	fmt.Printf("address:%+v\n", address)

	if len(address) == 0 {
		fmt.Println("请输入要查询的地址")
		return
	}
	totalBalance := client.Chain.GetBalance(address)
	fmt.Printf("用户%s的余额是%f\n", address, totalBalance)
}

func (client *Client) SendTransaction() {
	addnewblock := flag.NewFlagSet(SENDTRASACTION, flag.ExitOnError)
	from := addnewblock.String("from", "", "发起者地址")
	to := addnewblock.String("to", "", "接收者地址")
	value := addnewblock.String("value", "", "数值")
	// setcoinbase :=addnewblock.String("setcoinbase","","矿工地址")

	//labol :=addnewblock.String("labol","","数值")
	_ = addnewblock.Parse(os.Args[2:])
	//1.从参数中取出以 —开头的参数
	//2.准备一个当前命令支持的所有的参数切片
	err := client.Chain.SendTransaction(*from, *to, *value)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//	fmt.Println("交易成功")

}

func (client *Client) GetLastBlock() {
	set := os.Args[2:]
	if len(set) > 0 {
		fmt.Println("Error:unknow getlastblock mean")
		return
	}
	last := client.Chain.GetLastBlock()
	hashBig := new(big.Int).SetBytes(last.Hash[:])
	if hashBig.Cmp(big.NewInt(0)) > 0 {
		fmt.Println("最新区块高度:", last.Height)
		return
	}
	fmt.Println("还未曾有区块")
	fmt.Println("请使用go run main.go generategenesis生成区块")
}

func (client *Client) GetBlockCount() {
	blocks, err := client.Chain.GetAllBlocks()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("当前共有%d个区块", len(blocks))
}

func (client *Client) Default() {
	fmt.Println("不支持该命令功能")
	fmt.Println("请使用go run main.go help查看更多命令")
}

func (client *Client) GetAllBlocks() {
	if len(os.Args[2:]) > 0 {
		fmt.Println("getallblocks不接收参数")
		return
	}
	allblocks, err := client.Chain.GetAllBlocks()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, block := range allblocks {
		fmt.Printf("区块高度%d,区块hash：%x\n", block.Height, block.Hash)
		for index, tx := range block.Txs {
			fmt.Printf("区块%d的第%d交易,交易hash是：%x\n", block.Height, index, tx.TxHash)
			fmt.Println("\t该笔交易的交易输入")
			for inputidex, input := range tx.Inputs { //遍历交易输入
				fmt.Printf("\t\t第%d个交易输入，金额属于%x的第%d个\n", inputidex, input.Txid, input.Vout)
			}
			fmt.Println("\t该笔交易的交易输出")
			for indexoutput, output := range tx.Outputs { //交易输出
				fmt.Printf("\t\t第%d个交易输出，转给%x面额为%f的金额\n", indexoutput, output.PubHash, output.Value)
			}
		}
	}
}

/*
*

	用于想控制台输出项目的使用说明
*/
func (client *Client) Help() {

	fmt.Println("——————————————————Welcome To  PublicChain  NM coin————————————————————")
	fmt.Println()
	rand.Seed(time.Now().UnixNano())
	a := rand.Intn(2)
	if a == 1 {
		fmt.Println("                        .::::.\\m '+")
		fmt.Println("                      .::::::::.\\m' +")
		fmt.Println("                     :::::::::::  \\m' +")
		fmt.Println("                 ..:::::::::::'\\n' +")
		fmt.Println("               '::::::::::::'\\m' +")
		fmt.Println("                 .::::::::::\\m' +")
		fmt.Println("            '::::::::::::::..\\m' +")
		fmt.Println("                 ..::::::::::::.\\m' +")
		fmt.Println("               ``::::::::::::::::\\m' +")
		fmt.Println("                ::::``:::::::::'        .:::.\\' +")
		fmt.Println("               ::::'   ':::::'       .::::::::.\\m' +")
		fmt.Println("             .::::'      ::::     .:::::::'::::.\\m' +")
		fmt.Println("            .:::'       :::::  .:::::::::' ':::::.\\m' +")
		fmt.Println("           .::'        :::::.:::::::::'      ':::::.\\m' +")
		fmt.Println("          .::'         ::::::::::::::'         ``::::.\\m' +")
		fmt.Println("      ...:::           ::::::::::::'              ``::.\\m +")
		fmt.Println("     ````':.          ':::::::::'                  ::::..\\m' +")
		fmt.Println("                        '.:::::'                    ':'````..\\m' +")

	} else {
		fmt.Println("┌───┐   ┌───┬───┬───┬───┐ ┌───┬───┬───┬───┐ ┌───┬───┬───┬───┐ ┌───┬───┬───┐")
		fmt.Println("│Esc│   │ F1│ F2│ F3│ F4│ │ F5│ F6│ F7│ F8│ │ F9│F10│F11│F12│ │P/S│S L│P/B│")
		fmt.Println("└───┘   └───┴───┴───┴───┘ └───┴───┴───┴───┘ └───┴───┴───┴───┘ └───┴───┴───┘")
		fmt.Println("┌───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───────┐ ┌───┬───┬───┐ ┌───┬───┬───┬───┐")
		fmt.Println("│~ `│! 1│@ 2│# 3│$ 4│% 5│^ 6│& 7│* 8│( 9│) 0│_ -│+ =│ BacSp │ │Ins│Hom│PUp│ │N L│ / │ * │ - │")
		fmt.Println("├───┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─────┤ ├───┼───┼───┤ ├───┼───┼───┼───┤")
		fmt.Println("│ Tab │ Q │ W │ E │ R │ T │ Y │ U │ I │ O │ P │{ [│} ]│ |  \\│ │Del│End│PDn│ │ 7 │ 8 │ 9 |   │")
		fmt.Println("├─────┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴─────┤ └───┴───┴───┘ ├───┼───┼───┤ + │")
		fmt.Println("│ Caps │ A │ S │ D │ F │ G │ H │ J │ K │ L │: ;│\" '│ Enter │                │ 4 │ 5│ 6  |   │")
		fmt.Println("├──────┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴────────┤     ┌───┐     ├───┼───┼───┼───┤")
		fmt.Println("│ Shift  │ Z │ X │ C │ V │ B │ N │ M │< ,│> .│? /│  Shift   │     │ ↑ │     │ 1 │ 2 │ 3 │   │")
		fmt.Println("├─────┬──┴─┬─┴──┬┴───┴───┴───┴───┴───┴──┬┴───┼───┴┬────┬────┤ ┌───┼───┼───┐ ├───┴───┼───┤ E││")
		fmt.Println("│ Ctrl│Win │Alt │         Space         │ Alt│    │    │Ctrl│ │ ← │ ↓ │ → │ │   0   │ . │←─┘│")
		fmt.Println("└─────┴────┴────┴───────────────────────┴────┴────┴────┴────┘ └───┴───┴───┘ └───────┴───┴───┘")
	}

	fmt.Println()
	fmt.Println("使用説明：")
	fmt.Println("\tgo  run main.go command[arguments]")
	fmt.Println()
	fmt.Println("現在使用可能な説明:")
	fmt.Println()
	fmt.Println("\tThe commands are:")
	fmt.Println()
	fmt.Println("\t" + SENDTRASACTION + "\t\t\t 发送一笔交易-from -to -value")
	fmt.Println("\t" + GENERATEGENESIS + "\t\t\t 创建创世区块")
	fmt.Println("\t" + GETBLOCKCOUNT + "\t\t\t 捕获区块高度")
	fmt.Println("\t" + GETLASTBLOCK + "\t\t\t 获取最后一个区块")
	fmt.Println("\t" + GETALLBLOCKS + "\t\t\t 获取所有区块")
	fmt.Println("\t" + CREATECHAIN + "\t\t\t 创建区块")
	fmt.Println("\t" + GETNEWADDRESS + "\t\t\t 自动生成地址")
	fmt.Println("\t" + LISTADDRESS + "\t\t\t 查看所有地址列表")

	fmt.Println()
	fmt.Println("追加のヘルプタイプ")
}
