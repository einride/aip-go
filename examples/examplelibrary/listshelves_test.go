package examplelibrary

import (
	"context"
	"fmt"

	"google.golang.org/genproto/googleapis/example/library/v1"
)

func ExampleServer_ListShelves() {
	ctx := context.Background()
	server := &Server{
		Storage: &Storage{
			Shelves: []*library.Shelf{
				{Name: "shelves/0001", Theme: "Sci-Fi"},
				{Name: "shelves/0002", Theme: "Horror"},
				{Name: "shelves/0003", Theme: "Romance"},
			},
		},
	}
	page1, err := server.ListShelves(ctx, &library.ListShelvesRequest{
		PageSize: 2,
	})
	if err != nil {
		panic(err) // TODO: Handle errors.
	}
	for _, shelf := range page1.Shelves {
		fmt.Println(shelf.Name, shelf.Theme)
	}
	fmt.Println("page1.NextPageToken non-empty:", page1.NextPageToken != "")
	page2, err := server.ListShelves(ctx, &library.ListShelvesRequest{
		PageSize:  2,
		PageToken: page1.NextPageToken,
	})
	if err != nil {
		panic(err) // TODO: Handle errors.
	}
	for _, shelf := range page2.Shelves {
		fmt.Println(shelf.Name, shelf.Theme)
	}
	fmt.Println("page2.NextPageToken non-empty:", page2.NextPageToken != "")
	// Output:
	// shelves/0001 Sci-Fi
	// shelves/0002 Horror
	// page1.NextPageToken non-empty: true
	// shelves/0003 Romance
	// page2.NextPageToken non-empty: false
}
