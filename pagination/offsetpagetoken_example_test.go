package pagination

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// nolint: lll
func ExampleParseOffsetPageToken() {
	// This example simulates a database of 1000 books with the same parent (shelves/1).
	const exampleShelf = "shelves/1"
	books := make([]*library.Book, 1000)
	for i := range books {
		books[i] = &library.Book{Name: exampleShelf + "/books/" + strconv.Itoa(i)}
	}
	// This function simulates a request gRPC request handler for the ListBooks request.
	listBooksRequestHandler := func(
		ctx context.Context,
		request *library.ListBooksRequest,
	) (*library.ListBooksResponse, error) {
		// Apply defaults and coercion.
		if request.PageSize <= 0 {
			request.PageSize = 100
		}
		if request.PageSize > 1000 {
			request.PageSize = 1000
		}
		// Parse the page token.
		pageToken, err := ParseOffsetPageToken(request)
		if err != nil {
			fmt.Printf("failed to parse page token: %v", err)
			return nil, status.Error(codes.InvalidArgument, "invalid page token")
		}
		// Calculate page.
		pageStart := pageToken.Offset
		pageEnd := pageStart + int64(request.PageSize)
		// Return current page and next page token.
		switch {
		case request.Name != exampleShelf:
			return &library.ListBooksResponse{
				Books:         nil,
				NextPageToken: "",
			}, nil
		case pageStart >= int64(len(books)):
			return &library.ListBooksResponse{
				Books:         nil,
				NextPageToken: "",
			}, nil
		case pageEnd > int64(len(books)):
			return &library.ListBooksResponse{
				Books:         books[pageStart:],
				NextPageToken: "",
			}, nil
		default:
			return &library.ListBooksResponse{
				Books:         books[pageStart:pageEnd],
				NextPageToken: pageToken.Next(request).String(),
			}, nil
		}
	}
	output := func(m proto.Message) {
		fmt.Println(strings.ReplaceAll(protojson.MarshalOptions{}.Format(m), " ", ""))
	}
	request1 := &library.ListBooksRequest{
		Name:     exampleShelf,
		PageSize: 2,
	}
	output(request1)
	page1, err := listBooksRequestHandler(context.Background(), request1)
	if err != nil {
		panic(err)
	}
	for _, book := range page1.Books {
		output(book)
	}
	request2 := &library.ListBooksRequest{
		Name:      exampleShelf,
		PageSize:  3,
		PageToken: page1.NextPageToken,
	}
	output(request2)
	page2, err := listBooksRequestHandler(context.Background(), request2)
	if err != nil {
		panic(err)
	}
	for _, book := range page2.Books {
		output(book)
	}
	// Output:
	// {"name":"shelves/1","pageSize":2}
	// {"name":"shelves/1/books/0"}
	// {"name":"shelves/1/books/1"}
	// {"name":"shelves/1","pageSize":3,"pageToken":"PP-BAwEBD09mZnNldFBhZ2VUb2tlbgH_ggABAgEGT2Zmc2V0AQQAAQ9SZXF1ZXN0Q2hlY2tzdW0BBgAAAAv_ggEEAfzscdKdAA=="}
	// {"name":"shelves/1/books/2"}
	// {"name":"shelves/1/books/3"}
	// {"name":"shelves/1/books/4"}
}
