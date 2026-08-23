package pkg

import (
	"fmt"
	"strings"
)

// productNames maps a product's URL-facing name to its ProductList ID.
var productNames = map[string]ProductList{
	"synops": SynOps,
}

// ParseProductList resolves a product name (case-insensitive) to its ProductList ID.
func ParseProductList(name string) (ProductList, error) {
	product, ok := productNames[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("unknown product %q", name)
	}
	return product, nil
}
