package protoloader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func LoadFilesFromGoPackage(goPackage string) (*protoregistry.Files, error) {
	tmpDir, err := ioutil.TempDir(".", "protoloader*")
	if err != nil {
		return nil, fmt.Errorf("load proto files from Go package %s: %w", goPackage, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			panic(fmt.Errorf("failed to clean up temporary dir: %s", tmpDir))
		}
	}()
	filename := filepath.Join(tmpDir, "main.go")
	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("load proto files from Go package %s: %w", goPackage, err)
	}
	defer func() {
		_ = f.Close()
	}()
	if err := mainTemplate.Execute(f, struct{ GoPackage string }{GoPackage: goPackage}); err != nil {
		return nil, fmt.Errorf("load proto files from Go package %s: %w", goPackage, err)
	}
	cmd := exec.Command("go", "run", filename)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go run %s: %s", filename, stderr.String())
	}
	data, err := base64.StdEncoding.DecodeString(stdout.String())
	if err != nil {
		return nil, fmt.Errorf("load proto files from Go package %s: %w", goPackage, err)
	}
	var fileSet descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &fileSet); err != nil {
		return nil, fmt.Errorf("load proto files from Go package %s: %w", goPackage, err)
	}
	return protodesc.NewFiles(&fileSet)
}

// nolint: gochecknoglobals
var mainTemplate = template.Must(template.New("main").Parse(`
package main

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "{{.GoPackage}}" // package to load
)

func main() {
	fileSet := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, 0, protoregistry.GlobalFiles.NumFiles()),
	}
	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		fileSet.File = append(fileSet.File, protodesc.ToFileDescriptorProto(file))
		return true
	})
	data, err := proto.Marshal(fileSet)
	if err != nil {
		panic(err)
	}
	fmt.Print(base64.StdEncoding.EncodeToString(data))
}
`))
