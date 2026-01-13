package exception

import (
	"fmt"
	"github.com/TewApirat/food-shop/pkg/foodShop/domain"
)

type CouponSpendMinimumNotMetError struct {
	CouponCode   string
	MinimumSpend domain.Money
	CurrentSpend domain.Money
}

func (e *CouponSpendMinimumNotMetError) Error() string {
	return fmt.Sprintf("coupon %s requires minimum spend of %s, but current spend is %s", e.CouponCode, e.MinimumSpend, e.CurrentSpend)
}