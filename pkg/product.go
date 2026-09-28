package pkg

import (
	"fmt"
	"strings"
)

// productNames maps a product's URL-facing name to its ProductList ID.
var productNames = map[string]ProductList{
	"synops": SynOps,
}

// DashboardLink is the internal SPA route of a product's dashboard, e.g.
// "/dashboards/synops". It is also the APP.WebAppLinkList.Link a user must be
// granted (APP.User_WebAppLink) to access that product's data.
func DashboardLink(product ProductList) (string, bool) {
	for name, p := range productNames {
		if p == product {
			return "/dashboards/" + name, true
		}
	}
	return "", false
}

// ParseProductList resolves a product name (case-insensitive) to its ProductList ID.
func ParseProductList(name string) (ProductList, error) {
	product, ok := productNames[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("unknown product %q", name)
	}
	return product, nil
}
