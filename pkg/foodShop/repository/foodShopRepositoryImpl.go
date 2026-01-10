package repository

import (
	"sort"

	"github.com/TewApirat/food-shop/pkg/foodShop/domain"
	"github.com/TewApirat/food-shop/pkg/foodShop/exception"
	"github.com/TewApirat/food-shop/pkg/foodShop/model"
)

type foodShopRepositoryImpl struct {
	menu  map[model.MenuItemCode]model.MenuItem
	promo []model.Promotion
}

func NewFoodShopRepositoryImpl(
	menu map[model.MenuItemCode]model.MenuItem,
	promo []model.Promotion,
) FoodShopRepository {

	ownedMenu := make(map[model.MenuItemCode]model.MenuItem, len(menu))
	for code, menuItem := range menu {
		ownedMenu[code] = menuItem
	}

	ownedPromotions := make([]model.Promotion, len(promo))
	copy(ownedPromotions, promo)

	return &foodShopRepositoryImpl{
		menu:  ownedMenu,
		promo: ownedPromotions,
	}
}


func DefaultMenu() map[model.MenuItemCode]model.MenuItem {
	return map[model.MenuItemCode]model.MenuItem{
		"RED":    {Code: "RED", Name: "Red set", Price: domain.THB(50)},
		"GREEN":  {Code: "GREEN", Name: "Green set", Price: domain.THB(40)},
		"BLUE":   {Code: "BLUE", Name: "Blue set", Price: domain.THB(30)},
		"YELLOW": {Code: "YELLOW", Name: "Yellow set", Price: domain.THB(50)},
		"PINK":   {Code: "PINK", Name: "Pink set", Price: domain.THB(80)},
		"PURPLE": {Code: "PURPLE", Name: "Purple set", Price: domain.THB(90)},
		"ORANGE": {Code: "ORANGE", Name: "Orange set", Price: domain.THB(120)},
	}
}

func DefaultPromotions() []model.Promotion {
	return []model.Promotion{
		{
			Code:        "MEMBER",
			Title:       "Member card 10% off",
			Description: "Get 10% discount on the total bill if customer has a member card.",
		},
		{
			Code:        "PAIR",
			Title:       "Pair discount 5% (ORANGE/PINK/GREEN)",
			Description: "Every pair (2 items of the same code) for ORANGE/PINK/GREEN gets 5% off that pair value.",
		},
		{
			Code:        "OFF20",
			Title:       "Coupon OFF20 for 20 THB off",
			Description: "Get 20 THB off on total bill when using coupon code OFF20.",
		},
	}
}

func NewFoodShopRepositoryDefault() FoodShopRepository {
	return NewFoodShopRepositoryImpl(DefaultMenu(), DefaultPromotions())
}

func (r *foodShopRepositoryImpl) ListMenuItems() ([]model.MenuItem, error) {
	menuItems := make([]model.MenuItem, 0, len(r.menu))
	for _, menuItem := range r.menu {
		menuItems = append(menuItems, menuItem)
	}

	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].Code < menuItems[j].Code
	})

	return menuItems, nil
}

func (r *foodShopRepositoryImpl) FindMenuItemByCode(code model.MenuItemCode) (model.MenuItem, error) {
	menuItems, ok := r.menu[code]
	if !ok {
		return model.MenuItem{}, exception.UnknownMenuItemError{Code: string(code)}
	}
	return menuItems, nil
}

func (r *foodShopRepositoryImpl) ListPromotions() ([]model.Promotion, error) {
	promotions := make([]model.Promotion, len(r.promo))
	copy(promotions, r.promo)
	return promotions, nil
}
