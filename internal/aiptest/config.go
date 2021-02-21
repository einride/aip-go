package aiptest

type Config struct {
	Packages []PackageConfig `yaml:"packages"`
}

type PackageConfig struct {
	// Path contains the Go import path of the package to generate tests for.
	// Example: `go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1`
	Path string `yaml:"path"`
	// Services is the list of services to generate tests for.
	Services []ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Name string          `yaml:"name"`
	Out  GoPackageConfig `yaml:"out"`
}

type GoPackageConfig struct {
	// Name is the package name.
	Name string `yaml:"name"`
	// Path is the package import path.
	Path string `yaml:"path"`
}
