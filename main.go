package main

import (
	"fmt"
	"github.com/shopspring/decimal"
)

const (
	WeChat = 1
	AliPay  =2
	Balance=3
)

type Amount struct {
	ItemAmount float64
	OrderAmount float64
	WeChatAmount float64
	AliPayAmount float64
	BalanceAmount float64
	TaxAmount float64
}
/**
 计算税费
 billType: 1 普票
           2 增票
	税率默认：0.13
 */
func (a *Amount) ComputeTaxPrice(billType int8) {
	taxAmount:=float64(0)
	if billType==2{
		taxAmount, _ = decimal.NewFromFloat( a.ItemAmount* 0.13).Round(2).Float64()
	}
	a.TaxAmount=taxAmount
}

//计算订单金额
func (a *Amount) ComputeOrderPrice(payType int8) {
	orderAmount, _ := decimal.NewFromFloat(a.ItemAmount + a.TaxAmount).Round(2).Float64()
	a.OrderAmount=orderAmount
	switch  payType {
	case WeChat:
		a.WeChatAmount=orderAmount
	case AliPay:
		a.AliPayAmount=orderAmount
	case Balance:
		a.BalanceAmount=orderAmount
	}
}

func main() {
	//传入支付方式
	payType:=int8(1)
	//传入发票类型
	billType:=int8(2)
	//商品金额
	itemAmount:=float64(200)
	amountInfo :=Amount{ItemAmount: itemAmount}
	//计算税费 税费默认0.13
	amountInfo.ComputeTaxPrice(billType)
	//计算订单金额
	amountInfo.ComputeOrderPrice(payType)

	fmt.Printf("amountInfo :%+v \n",amountInfo)
}
