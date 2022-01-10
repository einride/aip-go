//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"go.einride.tech/mage-tools/mglog"
	"go.einride.tech/mage-tools/mgpath"
	"go.einride.tech/mage-tools/mgtool"
	"go.einride.tech/mage-tools/targets/mgapilinter"
	"go.einride.tech/mage-tools/targets/mgbuf"

	// mage:import proto
	"go.einride.tech/mage-tools/targets/mgclangformat"
)

type Proto mg.Namespace

func (Proto) All() {
	cleanup := mgpath.ChangeWorkDir(mgpath.FromGitRoot("proto"))
	defer cleanup()
	mg.SerialDeps(
		mg.F(mgclangformat.ClangFormatProto, "einride"),
		Proto.BufLint,
		Proto.ApiLinterLint,
		Proto.BufGenerate,
	)
}

func (Proto) BufLint() {
	mg.Deps(mgbuf.BufLint)
}

func (Proto) ProtocGenGo() error {
	logger := mglog.Logger("protoc-gen-go")
	newer, err := target.Glob("build/protoc-gen-go", ".../go.mod")
	if err != nil {
		return nil
	}
	if !newer {
		return nil
	}
	logger.Info("building binary...")
	return sh.Run("go", "build", "-o", "build/protoc-gen-go", "google.golang.org/protobuf/cmd/protoc-gen-go")
}

func (Proto) ProtocGenGoAip() error {
	logger := mglog.Logger("protoc-gen-go-aip")
	logger.Info("building binary...")
	return sh.Run("go", "build", "-o", "build/protoc-gen-go-aip", "../cmd/protoc-gen-go-aip")
}

func (Proto) BufGenerate(ctx context.Context) error {
	logger := mglog.Logger("buf-generate")
	ctx = logr.NewContext(ctx, logger)
	goGrpc, err := mgtool.GoInstall(ctx, "google.golang.org/grpc/cmd/protoc-gen-go-grpc", "v1.2.0")
	if err != nil {
		return err
	}
	mg.Deps(
		Proto.ProtocGenGo,
		Proto.ProtocGenGoAip,
	)
	path := fmt.Sprintf("%s:%s:%s", os.Getenv("PATH"), filepath.Dir(goGrpc), "../.mage/tools/protoc/3.15.7/bin")
	if err := os.Setenv("PATH", path); err != nil {
		return err
	}
	logger.Info("generating protobuf stubs...")
	return mgbuf.Buf(ctx, "generate", "--template", "buf.gen.yaml", "--path", "einride")
}

func (Proto) BuildDescriptor(ctx context.Context) error {
	logger := mglog.Logger("build/descriptor.pb")
	ctx = logr.NewContext(ctx, logger)
	protoFiles, err := mgpath.FindFilesWithExtension("einride", ".proto")
	if err != nil {
		return err
	}
	newer, err := target.Glob("build/descriptor.pb", protoFiles...)
	if err != nil {
		return err
	}
	if !newer {
		return nil
	}
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}
	logger.Info("generating proto descriptor...")
	return mgbuf.Buf(ctx, "build", "-o", "build/descriptor.pb")
}

func (Proto) ApiLinterLint(ctx context.Context) error {
	protoFiles, err := mgpath.FindFilesWithExtension("einride/example/freight", ".proto")
	if err != nil {
		return err
	}
	mg.CtxDeps(ctx, Proto.BuildDescriptor)
	args := []string{"--set-exit-status", "--config", "api-linter.yaml", "--descriptor-set-in", "build/descriptor.pb"}
	args = append(args, protoFiles...)
	return mgapilinter.APILinterLint(ctx, args...)
}
