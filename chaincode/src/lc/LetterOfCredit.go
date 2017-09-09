package main

import "time"



//信用证信息
type LCLetter struct {
	//L/C NO 信用证号(20字段)
	LCNo string `json:"lcNo"`
	//发报行
	SendBank Bank
	//收报行
	RecvBank Bank
	//开证申请人
	ApplyCorp Corporation
	//受益人
	BenefCorp Corporation
	//开证币种
	//currency string
	//开证金额
	Amount float64
	//承兑金额
	//acceptAmount float32
	//未付金额
	//notPayAmount float32
	//开证日期
	IssuseDate time.Time
	//到期日期
	ExpiryDate time.Time
	//货物描述
	//goods GoodsInfo
	//单据要求
	//Documents[] Document
	//交单期限，（天数）
	//presentPeriod int
	//保证金
	//DepositAmt float64
	////改证次数
	//amendTimes int
	////费用承担方
	//chargeTaker Corporation
	////到单次数
	//ABTimes int
	////是否有效
	//isValid bool
	////是否闭卷
	//isClose bool
	////是否撤销
	//isCancel bool
	//当前谁拥有这个信用证
	Owner LegalEntity
}

//银行
type Bank struct{
 	LegalEntity
}
//企业
type Corporation struct {
	LegalEntity
}
type LegalEntity struct{
	No string
	Name string
}
type Carrier struct{
	LegalEntity
}
//货运单
type BillofLanding struct{
	//货物编号
	GoodsNo string
	//货物描述
	GoodsDesc string
	//装船发运地
	LoadPortName string
	//目的地
	TransPortName string
	//最迟装船日
	LatestShipDate string
	//是否分批装运
	PartialShipment bool
	//物流号
	TrackingNo string
	//物流公司
	Carrier Carrier
	//实际发货时间
	ShippingTime time.Time
	//拥有者
	Owner LegalEntity

}
type Document struct {
	//文件的路径
	FileUri string
	//文件的Hash
	FileHash string
	//文件的签名
	Fileignature string
	//上传人
	Uploader string
}

