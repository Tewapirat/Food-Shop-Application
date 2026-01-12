package model

import "github.com/TewApirat/food-shop/pkg/foodShop/domain"


type PurchasingRequest struct {
	Items  map[string]int `json:"items"`
	Member bool           `json:"member"`
	CouponCode string    `json:"coupon_code"`
}

type OrderLine struct {
	Code      MenuItemCode
	Name      string
	Qty       int
	UnitPrice domain.Money
	LineTotal domain.Money
}

type OrderQuote struct {
	Lines          []OrderLine
	Subtotal       domain.Money
	PairDiscount   domain.Money
	CouponCodeDiscount domain.Money
	MemberDiscount domain.Money
	Total          domain.Money
}
