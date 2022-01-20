//go:build mage
// +build mage

package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
	"go.einride.tech/mage-tools/mglog"
	"go.einride.tech/mage-tools/mgpath"
	"go.einride.tech/mage-tools/mgtool"
	"go.einride.tech/mage-tools/tools/mgapilinter"
	"go.einride.tech/mage-tools/tools/mgbuf"
	"go.einride.tech/mage-tools/tools/mgclangformat"
)

type Proto mg.Namespace

func (Proto) All() {
	mg.SerialDeps(
		Proto.ClangFormatProto,
		Proto.BufLint,
		Proto.ApiLinterLint,
		Proto.BufGenerate,
	)
}

var protoDir = mgpath.FromGitRoot("proto")

func (Proto) BufLint(ctx context.Context) error {
	cmd := mgbuf.Command(ctx, "lint")
	cmd.Dir = protoDir
	return cmd.Run()
}

func (Proto) ClangFormatProto() error {
	cmd := mgclangformat.FormatProtoCommand(filepath.Join(protoDir, "einride"))
	cmd.Dir = protoDir
	return cmd.Run()
}

func (Proto) ProtocGenGo(ctx context.Context) error {
	_, err := mgtool.GoInstallWithModfile(
		ctx,
		"google.golang.org/protobuf/cmd/protoc-gen-go",
		mgpath.FromGitRoot("go.mod"),
	)
	return err
}

func (Proto) ProtocGenGoAip() error {
	cmd := mgtool.Command(
		"go",
		"build",
		"-o",
		filepath.Join(mgpath.Bins(), "protoc-gen-go-aip"),
		"../cmd/protoc-gen-go-aip",
	)
	cmd.Dir = protoDir
	mglog.Logger("protoc-gen-go-aip").Info("building binary...")
	return cmd.Run()
}

func (Proto) ProtocGenGoGrpc(ctx context.Context) error {
	_, err := mgtool.GoInstall(ctx, "google.golang.org/grpc/cmd/protoc-gen-go-grpc", "v1.2.0")
	return err
}

func (Proto) BufGenerate(ctx context.Context) error {
	mg.Deps(
		Proto.ProtocGenGo,
		Proto.ProtocGenGoGrpc,
		Proto.ProtocGenGoAip,
	)
	cmd := mgbuf.Command(ctx, "generate", "--template", "buf.gen.yaml", "--path", "einride")
	cmd.Dir = protoDir
	mglog.Logger("buf-generate").Info("generating protobuf stubs...")
	return cmd.Run()
}

func (Proto) BuildDescriptor(ctx context.Context) error {
	protoFiles, err := mgpath.FindFilesWithExtension(filepath.Join(protoDir, "einride"), ".proto")
	if err != nil {
		return err
	}
	newer, err := target.Glob(filepath.Join(protoDir, "build/descriptor.pb"), protoFiles...)
	if err != nil {
		return err
	}
	if !newer {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(protoDir, "build"), 0o755); err != nil {
		return err
	}
	cmd := mgbuf.Command(ctx, "build", "-o", "build/descriptor.pb")
	cmd.Dir = protoDir
	mglog.Logger("build/descriptor.pb").Info("generating proto descriptor...")
	return cmd.Run()
}

func (Proto) ApiLinterLint(ctx context.Context) error {
	protoFiles, err := mgpath.FindFilesWithExtension(filepath.Join(protoDir, "einride/example/freight"), ".proto")
	if err != nil {
		return err
	}
	mg.Deps(Proto.BuildDescriptor)
	args := []string{"--set-exit-status", "--config", "api-linter.yaml", "--descriptor-set-in", "build/descriptor.pb"}
	args = append(args, protoFiles...)
	cmd := mgapilinter.Command(ctx, args...)
	cmd.Dir = protoDir
	mglog.Logger("api-linter-lint").Info("linting gRPC APIs...")
	return cmd.Run()
}
