package main

import (
	"context"
	"io/fs"
	"path/filepath"

	"go.einride.tech/sage/sg"
	"go.einride.tech/sage/sgtool"
	"go.einride.tech/sage/tools/sgapilinter"
	"go.einride.tech/sage/tools/sgbuf"
	"go.einride.tech/sage/tools/sgclangformat"
)

type Proto sg.Namespace

func (Proto) All(ctx context.Context) error {
	sg.Deps(ctx, Proto.ClangFormatProto, Proto.BufLint)
	sg.Deps(ctx, Proto.APILinterLint, Proto.BufGenerate)
	sg.Deps(ctx, Proto.BufGenerateTestdata)
	return nil
}

func (Proto) BufLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting proto files...")
	cmd := sgbuf.Command(ctx, "lint")
	cmd.Dir = sg.FromGitRoot("proto")
	return cmd.Run()
}

func (Proto) ClangFormatProto(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting proto files...")
	var protoFiles []string
	if err := filepath.WalkDir(sg.FromGitRoot("proto"), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	}); err != nil {
		return err
	}
	return sgclangformat.FormatProtoCommand(ctx, protoFiles...).Run()
}

func (Proto) ProtocGenGo(ctx context.Context) error {
	sg.Logger(ctx).Println("installing...")
	_, err := sgtool.GoInstallWithModfile(ctx, "google.golang.org/protobuf/cmd/protoc-gen-go", sg.FromGitRoot("go.mod"))
	return err
}

func (Proto) ProtocGenGoGRPC(ctx context.Context) error {
	sg.Logger(ctx).Println("installing...")
	_, err := sgtool.GoInstall(ctx, "google.golang.org/grpc/cmd/protoc-gen-go-grpc", "v1.2.0")
	return err
}

func (Proto) ProtocGenGoAIP(ctx context.Context) error {
	sg.Logger(ctx).Println("building binary...")
	return sg.Command(ctx, "go", "build", "-o", sg.FromBinDir("protoc-gen-go-aip"), "./cmd/protoc-gen-go-aip").Run()
}

func (Proto) BufGenerate(ctx context.Context) error {
	sg.Deps(ctx, Proto.ProtocGenGo, Proto.ProtocGenGoGRPC, Proto.ProtocGenGoAIP)
	sg.Logger(ctx).Println("generating proto stubs...")
	cmd := sgbuf.Command(ctx, "generate", "--template", "buf.gen.yaml", "--path", "einride")
	cmd.Dir = sg.FromGitRoot("proto")
	return cmd.Run()
}

func (Proto) APILinterLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting gRPC APIs...")
	return sgapilinter.Run(ctx)
}

func (Proto) BufGenerateTestdata(ctx context.Context) error {
	sg.Deps(ctx, Proto.ProtocGenGoAIP)
	sg.Logger(ctx).Println("generating testdata stubs...")
	cmd := sgbuf.Command(ctx, "generate", "--path", "test")
	cmd.Dir = sg.FromGitRoot("cmd/protoc-gen-go-aip/internal/genaip/testdata")
	return cmd.Run()
}
