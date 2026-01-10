package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"

	"github.com/TewApirat/food-shop/pkg/foodShop/model"
	_foodShopService "github.com/TewApirat/food-shop/pkg/foodShop/service"
)

type FoodShopControllerImpl struct {
	in              io.ReadCloser
	out             io.Writer
	foodShopService _foodShopService.FoodShopService
}

func NewFoodShopControllerImpl(in io.ReadCloser, out io.Writer, foodShopService _foodShopService.FoodShopService) FoodShopController {
	return &FoodShopControllerImpl{
		in:              in,
		out:             out,
		foodShopService: foodShopService,
	}
}

func (c *FoodShopControllerImpl) ServeCLI() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "Select: ",
		Stdin:           c.in,
		Stdout:          c.out,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintln(c.out, "Readline init error:", err)
		return
	}
	defer rl.Close()

	for {
		fmt.Fprintln(c.out, "\n==== Food Shop CLI ====")
		fmt.Fprintln(c.out, "1) View all menu items")
		fmt.Fprintln(c.out, "2) View all promotions")
		fmt.Fprintln(c.out, "3) Quote order (JSON input)")
		fmt.Fprintln(c.out, "4) View order history")
		fmt.Fprintln(c.out, "0) Exit")

		rl.SetPrompt("Select: ")
		choice, err := readLine(rl)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(c.out, "\nEOF received. Thankyou.")
				return
			}
			fmt.Fprintln(c.out, "\nRead error:", err)
			return
		}
		if choice == "" {
			continue
		}

		switch choice {
		case "1":
			c.handleViewMenu()
		case "2":
			c.handleViewPromotions()
		case "3":
			if ok := c.handleQuoteOrderJSON(rl); !ok {
				return
			}
		case "4":
			c.handleViewOrderHistory()
		case "0":
			fmt.Fprintln(c.out, "Thankyou.")
			return
		default:
			fmt.Fprintln(c.out, "Invalid choice. Please select 0-3.")
		}
	}
}

func (c *FoodShopControllerImpl) handleViewMenu() {
	items, err := c.foodShopService.GetMenuCatalog()
	if err != nil {
		fmt.Fprintln(c.out, err)
		return
	}

	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "--- Menu Catalog ---")
	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "--------+--------------+-----------")
	fmt.Fprintln(c.out, "CODE    | NAME         | PRICE		")
	fmt.Fprintln(c.out, "--------+--------------+-----------")

	for _, it := range items {
		fmt.Fprintf(c.out, "%-7s | %-12s | %s\n", it.Code, it.Name, it.Price.String())
	}
}

func (c *FoodShopControllerImpl) handleViewPromotions() {
	promos, err := c.foodShopService.GetPromotions()
	if err != nil {
		fmt.Fprintln(c.out, err)
		return
	}

	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "--- Promotions ---")
	fmt.Fprintln(c.out)

	for _, p := range promos {
		fmt.Fprintf(c.out, "[%s] %s\n", p.Code, p.Title)
		fmt.Fprintf(c.out, " - %s\n", p.Description)
		fmt.Fprintln(c.out)
	}
}

func (c *FoodShopControllerImpl) handleQuoteOrderJSON(rl *readline.Instance) bool {
	fmt.Fprintln(c.out, "\nPaste order JSON in one line, then press Enter.")
	fmt.Fprintln(c.out, `Example: {"items":{"RED":1,"GREEN":2},"member":false}`)

	rl.SetPrompt("Order JSON: ")
	line, err := readLine(rl)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(c.out, "\nEOF received. Bye.")
			return false
		}
		fmt.Fprintln(c.out, "Read error:", err)
		return false
	}

	if strings.TrimSpace(line) == "" {
		fmt.Fprintln(c.out, "Error: empty input")
		return true
	}

	var req model.PurchasingRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		fmt.Fprintln(c.out, "Error: invalid JSON:", err)
		fmt.Fprintln(c.out, `Hint: {"items":{"RED":1,"GREEN":2},"member":false}`)
		return true
	}

	quote, err := c.foodShopService.QuoteOrder(req)
	if err != nil {
		fmt.Fprintln(c.out, err)
		return true
	}


	fmt.Fprintln(c.out, "\n--- Order Items ---")
	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "--------+--------------+-----+------------+-----------")
	fmt.Fprintln(c.out, "CODE    | NAME         | QTY | UNIT PRICE |   TOTAL.  ")
	fmt.Fprintln(c.out, "--------+--------------+-----+------------+-----------")

	for _, ln := range quote.Lines {
    	fmt.Fprintf(
        c.out,
        "%-7s | %-12s | %3d | %-10s | %s\n",
        ln.Code, ln.Name, ln.Qty, ln.UnitPrice.String(), ln.LineTotal.String(),
		)
	}

	fmt.Fprintln(c.out, "\n--- Order Quote ---")
	fmt.Fprintln(c.out)
	fmt.Fprintf(c.out, "%-16s : %s\n", "Subtotal",        quote.Subtotal.String())
	fmt.Fprintf(c.out, "%-16s : %s\n", "Pair Discount",   quote.PairDiscount.String())
	fmt.Fprintf(c.out, "%-16s : %s\n", "Member Discount", quote.MemberDiscount.String())
	fmt.Fprintf(c.out, "%-16s : %s\n", "Bulk Discount",   quote.BulkDiscount.String())
	fmt.Fprintf(c.out, "%-16s : %s\n", "Total",           quote.Total.String())	

	return true
}

func (c *FoodShopControllerImpl) handleViewOrderHistory() {
	count, err := c.foodShopService.CountOrderHistory()
	if err != nil {
		fmt.Fprintln(c.out, err)
		return
	}

	entries, err := c.foodShopService.ListOrderHistory()
	if err != nil {
		fmt.Fprintln(c.out, err)
		return
	}

	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "--- Order History ---")
	fmt.Fprintln(c.out)
	fmt.Fprintf(c.out, "Total orders: %d\n\n", count)

	if count == 0 {
		fmt.Fprintln(c.out, "No orders yet.")
		return
	}

	for _, e := range entries {
		fmt.Fprintf(c.out, "Order #%d | %s | member=%v\n",
			e.OrderNo, e.CreatedAt.Format("2006-01-02 15:04:05"), e.Member)
		fmt.Fprintln(c.out)

		fmt.Fprintln(c.out, "CODE    | NAME         | QTY | UNIT PRICE | LINE TOTAL")
		fmt.Fprintln(c.out, "--------+--------------+-----+------------+-----------")
		for _, ln := range e.Line {
			fmt.Fprintf(c.out, "%-7s | %-12s | %3d | %-10s | %s\n",
				ln.Code, ln.Name, ln.Qty, ln.UnitPrice.String(), ln.LineTotal.String())
		}

		fmt.Fprintln(c.out)
		fmt.Fprintf(c.out, "%-16s : %s\n", "Subtotal",        e.Subtotal.String())
		fmt.Fprintf(c.out, "%-16s : %s\n", "Pair Discount",   e.PairDiscount.String())
		fmt.Fprintf(c.out, "%-16s : %s\n", "Member Discount", e.MemberDiscount.String())
		fmt.Fprintf(c.out, "%-16s : %s\n", "Total",           e.Total.String())
		fmt.Fprintln(c.out, "\n------------------------------")
		fmt.Fprintln(c.out)
	}
}


// readline-aware readLine
func readLine(rl *readline.Instance) (string, error) {
	s, err := rl.Readline()
	if err != nil {
		// Ctrl+C
		if errors.Is(err, readline.ErrInterrupt) {
			return "", nil
		}
		// Ctrl+D => io.EOF
		return "", err
	}
	return strings.TrimSpace(s), nil
}
