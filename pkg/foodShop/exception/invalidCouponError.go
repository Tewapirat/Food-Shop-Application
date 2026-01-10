package exception

import "fmt"

type InvalidCouponError struct {
	Code string
}

func (e *InvalidCouponError) Error() string {
	return fmt.Sprintf("invalid coupon code: %s", e.Code)
}
