[![Go Reference](https://pkg.go.dev/badge/go.einride.tech/aip.svg)](https://pkg.go.dev/go.einride.tech/aip)

# AIP Go

Go SDK for implementing [Google API Improvement Proposals](https://aip.dev/)
(AIP).

## Documentation

See [https://aip.dev](https://aip.dev/) for the full AIP documentation.

## Installing

```bash
$ go get -u go.einride.tech/aip
```

## Examples

### [AIP-132](https://google.aip.dev/132) (Standard method: List)

- Use [`pagination.PageToken`](./pagination/pagetoken.go) to implement
  offset-based pagination.

  ```go
  package examplelibrary

  import (
      "context"

      "go.einride.tech/aip/pagination"
      "google.golang.org/genproto/googleapis/example/library/v1"
      "google.golang.org/grpc/codes"
      "google.golang.org/grpc/status"
  )

  func (s *Server) ListShelves(
      ctx context.Context,
      request *library.ListShelvesRequest,
  ) (*library.ListShelvesResponse, error) {
      // Handle request constraints.
      const (
  §           maxPageSize     = 1000
          defaultPageSize = 100
      )
      switch {
      case request.PageSize < 0:
          return nil, status.Errorf(codes.InvalidArgument, "page size is negative")
      case request.PageSize == 0:
          request.PageSize = defaultPageSize
      case request.PageSize > maxPageSize:
          request.PageSize = maxPageSize
      }
      // Use pagination.PageToken for offset-based page tokens.
      pageToken, err := pagination.ParsePageToken(request)
      if err != nil {
          return nil, status.Errorf(codes.InvalidArgument, "invalid page token")
      }
      // Query the storage.
      result, err := s.Storage.ListShelves(ctx, &ListShelvesQuery{
          Offset:   pageToken.Offset,
          PageSize: request.GetPageSize(),
      })
      if err != nil {
          return nil, err
      }
      // Build the response.
      response := &library.ListShelvesResponse{
          Shelves: result.Shelves,
      }
      // Set the next page token.
      if result.HasNextPage {
          response.NextPageToken = pageToken.Next(request).String()
      }
      // Respond.
      return response, nil
  }
  ```
