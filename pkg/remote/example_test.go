package remote_test

import (
	"context"
	"fmt"

	"github.com/bpineau/pofo/pkg/remote"
)

func ExampleService_Execute() {
	// Catalog search needs neither a data client nor network access.
	service := remote.New(nil, "example")
	result, err := service.Execute(context.Background(), "search_assets", []byte(`{"query":"IWDA","limit":1}`))
	if err != nil {
		panic(err)
	}
	page := result.Data.(remote.SearchData)
	fmt.Println(len(page.Items), page.Items[0].ISIN)
	// Output: 1 IE00B4L5Y983
}
