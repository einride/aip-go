# aiptest

Generates test fixtures for [AIP][google-aip] compliant services.

**Experimental**: This generator is under active development and breaking
changes to config files and generated code are expected between
releases.

[google-aip]: https://aip.dev/

## Usage

### Code generation

Use a YAML config file to specify the resources to generate tests for:

```yaml
packages:
  - path: go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1
    services:
      - name: FreightService
        out:
          name: freightsvc
          path: ./examples/freightsvc
```

### Running

```go
func Test_AIP(t *testing.T) {
	var ctx context.Context
	var service examplefreightv1.FreightServiceServer

	fx := &aipTestShipperFixture{
		Ctx:     ctx,
		Service: service,
		Create: func() *examplefreightv1.Shipper {
			return &examplefreightv1.Shipper{
				DisplayName: "Display name",
			}
		},
		Update: func() *examplefreightv1.Shipper {
			return &examplefreightv1.Shipper{
				DisplayName: "Updated display name",
			}
		},
		Skip: []string{
			// Skip all Get tests because [reasons].
			"Get",
		},
	}
	fx.Test(t)
}
```
