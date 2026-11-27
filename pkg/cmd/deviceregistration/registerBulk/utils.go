package registerBulk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices/registration"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// secret says which registration field a generated password goes to.
type secret int

const (
	noSecret      secret = iota // the device authenticates with its own certificate
	credentials                 // the device user's password
	enrollmentOTP               // the one-time password for the certificate request
)

// run registers each --id (or piped item) as one device with the bulk
// registration API and outputs the device's details. body builds the other
// properties from the flags and --data/--template.
func run(cmd *cobra.Command, f *cmdutil.Factory, authType string, password secret, body ...flags.GetOption) error {
	r, err := c8ystream.NewRunner(cmd, f)
	if err != nil {
		return err
	}
	// v2.59.0 sent the requests of the bulk registration commands without a page size
	r.NoDefaultPageSize()
	if err := r.InputFlag("id"); err != nil {
		return err
	}
	if err := r.Body(body...); err != nil {
		return cmderrors.NewUserError(err)
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	reg := client.Devices.Registration

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		raw, err := in.Body()
		if err != nil {
			return nil, err
		}
		item := gjson.ParseBytes(raw)
		externalID := in.String("id")
		if externalID == "" {
			externalID = item.Get("id").String()
		}
		opt := registration.RegisterOptions{
			ExternalID:     externalID,
			ExternalIDType: item.Get("external-type").String(),
			Name:           item.Get("name").String(),
			Type:           item.Get("type").String(),
			AuthType:       authType,
			ICCID:          item.Get("iccid").String(),
			Tenant:         item.Get("tenant").String(),
			GroupPath:      item.Get("group").String(),
			IsAgent:        item.Get("isAgent").String() == "true",
		}
		if opt.Name == "" {
			opt.Name = opt.ExternalID
		}
		if opt.ExternalIDType == "" {
			opt.ExternalIDType = registration.DefaultExternalIDType
		}

		// The output shows the password, as it is needed to connect the device.
		output := map[string]any{
			"externalId":   opt.ExternalID,
			"authType":     opt.AuthType,
			"name":         opt.Name,
			"type":         opt.Type,
			"externalType": opt.ExternalIDType,
			"tenant":       opt.Tenant,
		}
		if password != noSecret {
			value := item.Get("password").String()
			if value == "" {
				if value, err = reg.GeneratePassword(); err != nil {
					return nil, err
				}
			}
			if password == credentials {
				opt.Credentials = value
			} else {
				opt.EnrollmentOTP = value
			}
			output["password"] = value
		}
		tenant := opt.Tenant
		if tenant == "" {
			tenant = client.Auth.Tenant
		}
		output["username"] = fmt.Sprintf("%s/device_%s", tenant, opt.ExternalID)

		return func(ctx context.Context) any {
			return reg.Register(ctx, opt).Map[json.RawMessage](func(device jsonmodels.Identity) json.RawMessage {
				output["id"] = device.ManagedObjectID()
				b, _ := json.Marshal(output)
				return b
			})
		}, nil
	})
}
