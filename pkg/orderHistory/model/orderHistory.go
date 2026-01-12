package model

import (
	"time"

	"github.com/TewApirat/food-shop/pkg/foodShop/domain"
	"github.com/TewApirat/food-shop/pkg/foodShop/model"
	

)

type OrderHistoryEntry struct {
	OrderNo    int
	CreatedAt  time.Time
	Member     bool

	Line 		   []model.OrderLine
	Subtotal       domain.Money
	PairDiscount   domain.Money
	CouponCodeDiscount domain.Money
	MemberDiscount domain.Money
	Total          domain.Money
}
