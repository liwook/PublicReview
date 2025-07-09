package order

type voucher struct {
	ShopId      int    `json:"shopId"` //关联的商店id
	Title       string `json:"title"`
	SubTitle    string `json:"subTitle"`
	Rules       string `json:"rules"`
	PayValue    int    `json:"payValue"` //优惠的价格
	ActualValue int    `json:"actualValue"`
	Type        int    `json:"type"`  //优惠卷类型
	Stock       int    `json:"stock"` //库存
	BeginTime   string `json:"beginTime"`
	EndTime     string `json:"endTime"`
}

type seckillResquest struct {
	VoucherId int `json:"voucherId"`
	UserId    int `json:"userId"`
}
