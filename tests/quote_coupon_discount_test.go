package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/TewApirat/food-shop/pkg/foodShop/domain"
	_foodShopException "github.com/TewApirat/food-shop/pkg/foodShop/exception"
	_foodShopModel "github.com/TewApirat/food-shop/pkg/foodShop/model"
	_foodShopRepository "github.com/TewApirat/food-shop/pkg/foodShop/repository"
	_foodShopService "github.com/TewApirat/food-shop/pkg/foodShop/service"
	_orderHistoryModel "github.com/TewApirat/food-shop/pkg/orderHistory/model"
	_orderHistoryRepository "github.com/TewApirat/food-shop/pkg/orderHistory/repository"
)


func TestQuoteOrder_Coupon_Success(t *testing.T) {
	type tc struct {
		label         string
		in            _foodShopModel.PurchasingRequest
		expected      _foodShopModel.OrderQuote
		expectedQty   map[_foodShopModel.MenuItemCode]int
		setupMenuMock func(r *_foodShopRepository.FoodShopRepositoryMock)
	}

	cases := []tc{
		{
			label: "Success: OFF30 applies when subtotal >= 1000 (ORANGE 9), no member",
			in: _foodShopModel.PurchasingRequest{
				Items:      map[string]int{"ORANGE": 9},
				Member:     false,
				CouponCode: "OFF30",
			},
			expected: _foodShopModel.OrderQuote{
				Subtotal:           domain.THB(1080),
				PairDiscount:       domain.THB(48),
				CouponCodeDiscount: satang(30960), // 309.60 (30% of 1032.00)
				MemberDiscount:     domain.THB(0),
				Total:              satang(72240), // 722.40
			},
			expectedQty: map[_foodShopModel.MenuItemCode]int{"ORANGE": 9},
			setupMenuMock: func(r *_foodShopRepository.FoodShopRepositoryMock) {
				r.On("FindMenuItemByCode", _foodShopModel.MenuItemCode("ORANGE")).
					Return(_foodShopModel.MenuItem{Code: "ORANGE", Name: "Orange set", Price: domain.THB(120)}, nil).
					Once()
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.label, func(t *testing.T) {
			foodShopRepositoryMock := new(_foodShopRepository.FoodShopRepositoryMock)
			orderHistoryRepositoryMock := new(_orderHistoryRepository.OrderHistoryRepositoryMock)

			foodShopRepositoryMock.Test(t)
			orderHistoryRepositoryMock.Test(t)

			c.setupMenuMock(foodShopRepositoryMock)

			orderHistoryRepositoryMock.
				On("Add", mock.MatchedBy(func(entry _orderHistoryModel.OrderHistoryEntry) bool {
					// เวลา/เลขออเดอร์ ตรวจแบบหลวม ๆ ให้เสถียร
					if entry.OrderNo != 1 {
						return false
					}
					if entry.CreatedAt.IsZero() {
						return false
					}
					if time.Since(entry.CreatedAt) > time.Minute {
						return false
					}

					if entry.Member != c.in.Member {
						return false
					}
					if entry.Subtotal != c.expected.Subtotal {
						return false
					}
					if entry.PairDiscount != c.expected.PairDiscount {
						return false
					}
					if entry.CouponCodeDiscount != c.expected.CouponCodeDiscount {
						return false
					}
					if entry.MemberDiscount != c.expected.MemberDiscount {
						return false
					}
					if entry.Total != c.expected.Total {
						return false
					}

					gotQty := lineQtyMap(entry.Line)
					if len(gotQty) != len(c.expectedQty) {
						return false
					}
					for code, qty := range c.expectedQty {
						if gotQty[code] != qty {
							return false
						}
					}
					return true
				})).
				Return(nil).
				Once()

			svc := _foodShopService.NewFoodShopServiceImpl(
				foodShopRepositoryMock,
				orderHistoryRepositoryMock,
			)

			result, err := svc.QuoteOrder(c.in)
			assert.NoError(t, err)

			assert.Equal(t, c.expected.Subtotal, result.Subtotal)
			assert.Equal(t, c.expected.PairDiscount, result.PairDiscount)
			assert.Equal(t, c.expected.CouponCodeDiscount, result.CouponCodeDiscount)
			assert.Equal(t, c.expected.MemberDiscount, result.MemberDiscount)
			assert.Equal(t, c.expected.Total, result.Total)

			foodShopRepositoryMock.AssertExpectations(t)
			orderHistoryRepositoryMock.AssertExpectations(t)
		})
	}
}

func TestQuoteOrder_Coupon_Fail(t *testing.T) {
	type tc struct {
		label         string
		in            _foodShopModel.PurchasingRequest
		setupMenuMock func(r *_foodShopRepository.FoodShopRepositoryMock)
		check         func(t *testing.T, err error)
	}

	cases := []tc{
		{
			label: "Fail: OFF30 but subtotal < 1000 => minimum spend not met",
			in: _foodShopModel.PurchasingRequest{
				Items:      map[string]int{"RED": 1, "GREEN": 2}, // 130
				Member:     false,
				CouponCode: "OFF30",
			},
			setupMenuMock: func(r *_foodShopRepository.FoodShopRepositoryMock) {
				r.On("FindMenuItemByCode", _foodShopModel.MenuItemCode("RED")).
					Return(_foodShopModel.MenuItem{Code: "RED", Name: "Red set", Price: domain.THB(50)}, nil).
					Once()
				r.On("FindMenuItemByCode", _foodShopModel.MenuItemCode("GREEN")).
					Return(_foodShopModel.MenuItem{Code: "GREEN", Name: "Green set", Price: domain.THB(40)}, nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				var target *_foodShopException.CouponSpendMinimumNotMetError
				assert.True(t, errors.As(err, &target))
				assert.Equal(t, "OFF30", target.CouponCode)
				assert.Equal(t, domain.THB(1000), target.MinimumSpend)
				assert.Equal(t, domain.THB(130), target.CurrentSpend) // ต้องเป็น subtotal
			},
		},
		{
			label: "Fail: invalid coupon code => InvalidCouponError",
			in: _foodShopModel.PurchasingRequest{
				Items:      map[string]int{"RED": 1},
				Member:     false,
				CouponCode: "NOPE",
			},
			setupMenuMock: func(r *_foodShopRepository.FoodShopRepositoryMock) {
				r.On("FindMenuItemByCode", _foodShopModel.MenuItemCode("RED")).
					Return(_foodShopModel.MenuItem{Code: "RED", Name: "Red set", Price: domain.THB(50)}, nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				var target *_foodShopException.InvalidCouponError
				assert.True(t, errors.As(err, &target))
				assert.Equal(t, "NOPE", target.Code)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.label, func(t *testing.T) {
			foodShopRepositoryMock := new(_foodShopRepository.FoodShopRepositoryMock)
			orderHistoryRepositoryMock := new(_orderHistoryRepository.OrderHistoryRepositoryMock)

			foodShopRepositoryMock.Test(t)
			orderHistoryRepositoryMock.Test(t)

			c.setupMenuMock(foodShopRepositoryMock)

			svc := _foodShopService.NewFoodShopServiceImpl(
				foodShopRepositoryMock,
				orderHistoryRepositoryMock,
			)

			res, err := svc.QuoteOrder(c.in)
			assert.Equal(t, _foodShopModel.OrderQuote{}, res)
			assert.Error(t, err)
			c.check(t, err)

			// coupon fail ต้องไม่บันทึก history
			orderHistoryRepositoryMock.AssertNotCalled(t, "Add", mock.Anything)

			foodShopRepositoryMock.AssertExpectations(t)
			orderHistoryRepositoryMock.AssertExpectations(t)
		})
	}
}
