package freightsvc

import (
	"context"
	"testing"

	examplefreightv1 "go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1"
)

func Test_AIP(t *testing.T) {
	t.Skip("This is just an example.")
	ctx := context.Background()
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
