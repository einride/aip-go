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
	"go.einride.tech/mage-tools/mgmake"
	"go.einride.tech/mage-tools/mgpath"
	"go.einride.tech/mage-tools/mgtool"
	"go.einride.tech/mage-tools/targets/mgbuf"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggo"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggolangcilint"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggoreview"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgmarkdownfmt"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgconvco"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggitverifynodiff"
)

func init() {
	mgmake.GenerateMakefiles(
		mgmake.Makefile{
			Path:          mgpath.FromGitRoot("Makefile"),
			DefaultTarget: All,
		},
		mgmake.Makefile{
			Path:          mgpath.FromGitRoot("proto/Makefile"),
			DefaultTarget: Proto.All,
			Namespace:     Proto{},
		},
	)
}

func All() {
	mg.Deps(
		mg.F(mgconvco.ConvcoCheck, "origin/master..HEAD"),
		mgmarkdownfmt.FormatMarkdown,
		GoStringer,
		Proto.All,
	)
	mg.Deps(
		mggolangcilint.GolangciLint,
		mggoreview.Goreview,
		mggo.GoTest,
	)
	mg.SerialDeps(
		mggo.GoModTidy,
		mggitverifynodiff.GitVerifyNoDiff,
	)
}

func ProtocGenGoAip() error {
	logger := mglog.Logger("protoc-gen-go-aip")
	logger.Info("building binary...")
	return sh.Run("go", "build", "-o", "build/protoc-gen-go-aip", "./cmd/protoc-gen-go-aip")
}

func BufGenerateTestdata(ctx context.Context) error {
	logger := mglog.Logger("buf")
	ctx = logr.NewContext(ctx, logger)
	mg.SerialDeps(ProtocGenGoAip)
	logger.Info("generating testdata stubs...")
	cleanup := mgpath.ChangeWorkDir("cmd/protoc-gen-go-aip/internal/genaip/testdata")
	defer cleanup()
	return mgbuf.Buf(ctx, "generate", "--path", "test")
}

func GoStringer(ctx context.Context) error {
	logger := mglog.Logger("go-stringer")
	ctx = logr.NewContext(ctx, logger)
	goStringer, err := mgtool.GoInstall(ctx, "golang.org/x/tools/cmd/stringer", "v0.1.8")
	if err != nil {
		return err
	}
	methodTypeDir := mgpath.FromGitRoot("reflect/aipreflect")
	methodType := filepath.Join(methodTypeDir, "methodtype.go")
	methodTypeString := filepath.Join(methodTypeDir, "methodtype_string.go")
	generate, err := target.Glob(methodTypeString, methodType)
	if err != nil {
		return err
	}
	if generate {
		logger.Info("generating", "src", methodType, "dst", methodTypeString)
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Dir(goStringer))
		err = sh.RunWith(map[string]string{"PATH": path}, "go", "generate", methodType)
		if err != nil {
			return err
		}
	}
	return nil
}
