package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/TewApirat/food-shop/pkg/foodShop/domain"
	_foodShopException "github.com/TewApirat/food-shop/pkg/foodShop/exception"
	_foodShopModel "github.com/TewApirat/food-shop/pkg/foodShop/model"
	_foodShopRepository "github.com/TewApirat/food-shop/pkg/foodShop/repository"
	_orderHistoryModel "github.com/TewApirat/food-shop/pkg/orderHistory/model"
	_orderHistoryReppsitory "github.com/TewApirat/food-shop/pkg/orderHistory/repository"
)

var pairDiscountEligibleCodes = map[_foodShopModel.MenuItemCode]bool{
	"ORANGE": true,
	"PINK":   true,
	"GREEN":  true,
}

const (
	pairBundleSize          = 2
	pairDiscountPercent     = 5
	memberDiscountPercent   = 10
)

type foodShopServiceImpl struct {
	foodShopRepository _foodShopRepository.FoodShopRepository
	orderHistoryRepository _orderHistoryReppsitory.OrderHistoryRepository
	orderNo int
}

func NewFoodShopServiceImpl(
	foodShopRepository _foodShopRepository.FoodShopRepository,
	orderHistoryRepository _orderHistoryReppsitory.OrderHistoryRepository,
	) FoodShopService {
	return &foodShopServiceImpl{
		foodShopRepository: foodShopRepository,
		orderHistoryRepository: orderHistoryRepository,
	}
}

func (s *foodShopServiceImpl) ListOrderHistory() ([]_orderHistoryModel.OrderHistoryEntry, error) {
	return s.orderHistoryRepository.List()
}

func (s *foodShopServiceImpl) CountOrderHistory() (int, error) {
	return s.orderHistoryRepository.Count()
}


func (s *foodShopServiceImpl) GetMenuCatalog() ([]_foodShopModel.MenuItem, error) {
	return s.foodShopRepository.ListMenuItems()
}

func (s *foodShopServiceImpl) GetPromotions() ([]_foodShopModel.Promotion, error) {
	return s.foodShopRepository.ListPromotions()
}

// QuoteOrder workflow
// 1) Validate request
// 2) Prepare state for calculation / promotion rules
// 3) Process each input item (rawCode -> qty)
// 4) Calculate pair discount (policy-level discount)
// 5) Apply discounts in correct order
// 6) Persist order history (side effect)
// 7) Return quote result



func (s *foodShopServiceImpl) QuoteOrder(req _foodShopModel.PurchasingRequest) (_foodShopModel.OrderQuote, error) {
	if len(req.Items) == 0 {
		return _foodShopModel.OrderQuote{}, &_foodShopException.EmptyOrderError{}
	}

	qtyByCode := make(map[_foodShopModel.MenuItemCode]int)
	priceByCode := make(map[_foodShopModel.MenuItemCode]domain.Money)



	lines := make([]_foodShopModel.OrderLine, 0, len(req.Items))

	var subtotal domain.Money

	for rawCode, qty := range req.Items {
		if qty < 1 {
			return _foodShopModel.OrderQuote{}, &_foodShopException.InvalidQuantityError{Qty: qty}
		}

		code, err := normalizeItemCode(rawCode)
		if err != nil {
			return _foodShopModel.OrderQuote{}, err
		}

		menuItem, err := s.foodShopRepository.FindMenuItemByCode(code)
		if err != nil {
			return _foodShopModel.OrderQuote{}, fmt.Errorf("find menu item by code %s: %w", code, err)
		}

		priceByCode[code] = menuItem.Price
		qtyByCode[code] += qty




		lineTotal := menuItem.Price.MulInt(qty)
		subtotal = subtotal.Add(menuItem.Price.MulInt(qty))


		lines = append(lines, _foodShopModel.OrderLine{
			Code:      code,
			Name:      menuItem.Name,
			Qty:       qty,
			UnitPrice: menuItem.Price,
			LineTotal: lineTotal,
		})
	}

	pairDiscount, err := calculatePairDiscount(qtyByCode, priceByCode)
	if err != nil {
		return _foodShopModel.OrderQuote{}, err
	}

	afterPairDiscount := subtotal.Sub(pairDiscount)
	conponDiscount, err := calculateCouponDiscount(afterPairDiscount,req.Coupon)
	if err != nil {
		return _foodShopModel.OrderQuote{}, err
	}
	afterCouponDiscount := afterPairDiscount.Sub(conponDiscount)


	var memberDiscount domain.Money
	if req.Member {
		memberDiscount = afterCouponDiscount.Percent(memberDiscountPercent)
	}

	total := afterCouponDiscount.Sub(memberDiscount)

	s.orderNo++

	_ = s.orderHistoryRepository.Add(_orderHistoryModel.OrderHistoryEntry{
	OrderNo:        s.orderNo,
	CreatedAt:      time.Now(),
	Member:         req.Member,
	Line:          lines,
	Subtotal:       subtotal,
	PairDiscount:   pairDiscount,
	Coupon:	  conponDiscount,
	MemberDiscount: memberDiscount,
	Total:          total,
	
})


	return _foodShopModel.OrderQuote{
		Lines:          lines,
		Subtotal:       subtotal,
		PairDiscount:   pairDiscount,
		Coupon:	  conponDiscount,
		MemberDiscount: memberDiscount,
		Total:          total,
	}, nil
}

func normalizeItemCode(raw string) (_foodShopModel.MenuItemCode, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", &_foodShopException.InvalidItemCodeError{}
	}
	return _foodShopModel.MenuItemCode(strings.ToUpper(trimmed)), nil
}


func calculatePairDiscount(
	qtyByCode map[_foodShopModel.MenuItemCode]int,
	priceByCode map[_foodShopModel.MenuItemCode]domain.Money,
) (domain.Money, error) {

	totalDiscount := domain.Money(0)

	for code, qty := range qtyByCode {
		if !pairDiscountEligibleCodes[code] || qty < pairBundleSize {
			continue
		}

		unitPrice, ok := priceByCode[code]
		if !ok {
			return domain.Money(0), &_foodShopException.MenuItemPriceMissingError{Code: code}
		}

		pairCount := qty / pairBundleSize
		bundleValue := unitPrice.MulInt(pairBundleSize)
		discountPerBundle := bundleValue.Percent(pairDiscountPercent)

		totalDiscount = totalDiscount.Add(discountPerBundle.MulInt(pairCount))
	}

	return totalDiscount, nil
}

func calculateCouponDiscount(base domain.Money, couponCode string) (domain.Money, error) {
	code := strings.ToUpper(strings.TrimSpace(couponCode))

	if code == "" {
		return domain.THB(0), nil
	}

	switch code {
	case "OFF20":
		minBase := domain.THB(200)
		discount := domain.THB(20)
		if base >= minBase{
			return discount, nil
		}
		return domain.THB(0), nil
	default:
		return domain.THB(0), &_foodShopException.InvalidCouponError{Code: couponCode}
	}
}
