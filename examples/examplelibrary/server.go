package examplelibrary

import "google.golang.org/genproto/googleapis/example/library/v1"

type Server struct {
	library.UnimplementedLibraryServiceServer
	Storage *Storage
}
