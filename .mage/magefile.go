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
	"github.com/magefile/mage/target"
	"go.einride.tech/mage-tools/mglog"
	"go.einride.tech/mage-tools/mgmake"
	"go.einride.tech/mage-tools/mgpath"
	"go.einride.tech/mage-tools/mgtool"
	"go.einride.tech/mage-tools/targets/mggitverifynodiff"
	"go.einride.tech/mage-tools/targets/mgyamlfmt"
	"go.einride.tech/mage-tools/tools/mgbuf"
	"go.einride.tech/mage-tools/tools/mgconvco"
	"go.einride.tech/mage-tools/tools/mggo"
	"go.einride.tech/mage-tools/tools/mggolangcilint"
	"go.einride.tech/mage-tools/tools/mggoreview"
	"go.einride.tech/mage-tools/tools/mgmarkdownfmt"
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
		ConvcoCheck,
		FormatMarkdown,
		mgyamlfmt.FormatYAML,
		GoStringer,
		Proto.All,
	)
	mg.Deps(
		GolangciLint,
		Goreview,
		GoTest,
	)
	mg.SerialDeps(
		GoModTidy,
		mggitverifynodiff.GitVerifyNoDiff,
	)
}

func ConvcoCheck(ctx context.Context) error {
	mglog.Logger("convco-check").Info("checking...")
	return mgconvco.Command(ctx, "check", "origin/master..HEAD").Run()
}

func FormatMarkdown(ctx context.Context) error {
	mglog.Logger("format-markdown").Info("formatting..")
	return mgmarkdownfmt.Command(ctx, "-w", ".").Run()
}

func GolangciLint(ctx context.Context) error {
	mglog.Logger("golangci-lint").Info("running...")
	return mggolangcilint.LintCommand(ctx).Run()
}

func Goreview(ctx context.Context) error {
	mglog.Logger("goreview").Info("running...")
	return mggoreview.Command(ctx, "-c", "1", "./...").Run()
}

func GoModTidy() error {
	mglog.Logger("go-mod-tidy").Info("tidying Go module files...")
	return mggo.GoModTidy().Run()
}

func GoTest() error {
	mglog.Logger("go-test").Info("running Go unit tests..")
	return mggo.GoTest().Run()
}

func ProtocGenGoAip() error {
	mglog.Logger("protoc-gen-go-aip").Info("building binary...")
	return mgtool.Command("go", "build", "-o", "build/protoc-gen-go-aip", "./cmd/protoc-gen-go-aip").Run()
}

func BufGenerateTestdata(ctx context.Context) error {
	mg.SerialDeps(ProtocGenGoAip)
	cmd := mgbuf.Command(ctx, "generate", "--path", "test")
	cmd.Dir = mgpath.FromGitRoot("cmd/protoc-gen-go-aip/internal/genaip/testdata")
	mglog.Logger("buf").Info("generating testdata stubs...")
	return cmd.Run()
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
		cmd := mgtool.Command("go", "generate", methodType)
		cmd.Env = append(cmd.Env, "PATH="+path)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
