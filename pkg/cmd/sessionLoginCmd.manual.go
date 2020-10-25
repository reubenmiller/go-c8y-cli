package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal/v3"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sessionLoginCmd struct {
	TFACode  string
	LoginErr error
	LoginOK  bool

	*baseCmd
}

func newSessionLoginCmd() *sessionLoginCmd {
	ccmd := &sessionLoginCmd{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login into a cumulocity session",
		Long:  `Login and test the Cumulocity session and get either OAuth2 token, or using two factor authentication`,
		Example: `
c8y session login

Log into the current session
		`,
		RunE: ccmd.login,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	ccmd.bindEnv()

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *sessionLoginCmd) bindEnv() {
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// bindEnv("passphrase", "")
}

func (n *sessionLoginCmd) login(cmd *cobra.Command, args []string) error {
	if n.TFACode == "" {
		// read tfa code from env variable (if present)
		if tfaCode := os.Getenv("C8Y_TFA_CODE"); tfaCode != "" {
			n.TFACode = tfaCode
		}
	}

	loginOptions, _, err := client.Tenant.GetLoginOptions(context.Background())

	if err != nil {
		return fmt.Errorf("could not get login options: %w", err)
	}

	requiresTFA := false
	optionOrder := map[string]int{
		c8y.AuthMethodBasic:          2,
		c8y.AuthMethodOAuth2Internal: 1,
	}

	// sort login options
	sort.SliceStable(loginOptions.LoginOptions[:], func(i, j int) bool {
		iWeight := 100
		jWeight := 200

		if v, ok := optionOrder[loginOptions.LoginOptions[i].Type]; ok {
			iWeight = v
		}

		if v, ok := optionOrder[loginOptions.LoginOptions[j].Type]; ok {
			jWeight = v
		}
		return iWeight < jWeight
	})

	var currentTenant *c8y.CurrentTenant

	accessOK := false

	// 1. check existing cookies
	if len(client.Cookies) > 0 {
		// tenant, resp, err := client.Tenant.GetCurrentTenant(context.Background())
		_, resp, err := client.User.GetCurrentUser(context.Background())

		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			if v, ok := err.(*c8y.ErrorResponse); ok {
				Logger.Debug("Token is invalid, clearing cookies. err=%s", v.Message)
				client.SetCookies([]*http.Cookie{})
				WriteAuth(viper.GetViper())
			}
		} else {
			accessOK = true
		}
	}

	// 2. check if TFA is required
	if !accessOK {
		tenant, resp, err := client.Tenant.GetCurrentTenant(context.Background())

		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			if v, ok := err.(*c8y.ErrorResponse); ok {

				if strings.Contains(v.Message, "TFA TOTP setup required") {
					if err := n.setupTFA(); err != nil {
						return err
					}
					requiresTFA = true
				} else if strings.Contains(v.Message, "TFA TOTP code required") {
					Logger.Debug("TFA code is required. server response: %s", v.Message)
					requiresTFA = true
				}
			}
		} else {
			currentTenant = tenant
		}
	}

	// iterate through login options
	for _, option := range loginOptions.LoginOptions {
		switch option.Type {
		case c8y.AuthMethodOAuth2Internal:

			if requiresTFA && option.TFAStrategy == "TOTP" {
				if n.TFACode == "" {
					prompt := promptui.Prompt{
						Stdin:  os.Stdin,
						Stdout: os.Stderr,
						Label:  "Enter Two-Factor code",
						Validate: func(input string) error {
							if len(strings.ReplaceAll(input, " ", "")) < 6 {
								return fmt.Errorf("Non-zero input")
							}

							client.TFACode = input

							ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
							defer cancel()

							if err := client.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
								Logger.Errorf("OAuth2 failed. %s", err)
								return err
							}
							n.LoginOK = true
							return nil
						},
					}

					if v, err := prompt.Run(); err == nil {
						n.TFACode = v
					}
				}
				client.TFACode = n.TFACode
			}

			if !n.LoginOK {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
				defer cancel()

				Logger.Debugf("Logging in using %s", c8y.AuthMethodOAuth2Internal)
				if err := client.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
					Logger.Errorf("OAuth2 failed. %s", err)
					continue
				}
			}

		case c8y.AuthMethodBasic:
			// do nothing
		}

		// Verify credentials
		currentTenant, _, err = client.Tenant.GetCurrentTenant(context.Background())
		if err != nil {
			Logger.Infof("Could not get current tenant info. %s", err)
			continue
		}
		break
	}

	if currentTenant == nil || currentTenant.Name == "" {
		return fmt.Errorf("could not get current tenant info")
	}

	cliConfig.Persistent.Set("tenant", currentTenant.Name)
	Logger.Infof("Tenant: %s", currentTenant.Name)

	Logger.Infof("login2 cookies. %v", client.Cookies)

	return WriteAuth(viper.GetViper())
}

func (n *sessionLoginCmd) setupTFA() error {

	// Request TFA secret
	backupAuthMethod := client.AuthorizationMethod
	client.AuthorizationMethod = c8y.AuthMethodBasic
	resp, err := client.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret",
			DryRun: globalFlagDryRun,
		})

	if err != nil {
		Logger.Infof("Could not get tot")
		return err
	}

	// Display TOTP secret
	if v := resp.JSON.Get("rawSecret"); v.Exists() {
		totpURL := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", client.Username, v.String(), client.BaseURL.Host)
		qrterminal.GenerateWithConfig(totpURL, qrterminal.Config{
			Level:     qrterminal.M,
			Writer:    n.cmd.ErrOrStderr(),
			BlackChar: qrterminal.BLACK,
			WhiteChar: qrterminal.WHITE,
			QuietZone: 1,
		})

		n.cmd.Printf("\nTOTP Secret: %s\n\n", v.String())
	}

	// Verify TOTP by checking a code
	tfaCodePrompt := promptui.Prompt{
		Stdin:    os.Stdin,
		Stdout:   os.Stderr,
		Label:    "Enter Two-Factor code",
		Validate: n.verifyTFASetupCode,
	}

	if _, err := tfaCodePrompt.Run(); err != nil {
		return err
	}

	// Activate totp
	resp, err = client.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret/activity",
			Body:   map[string]interface{}{"isActive": true},
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to activate TFA (TOTP): %w", err)
	}

	time.Sleep(1000 * time.Millisecond)

	client.AuthorizationMethod = backupAuthMethod
	return nil
}

func (n sessionLoginCmd) verifyTFASetupCode(input string) error {
	if len(strings.ReplaceAll(input, " ", "")) < 6 {
		return fmt.Errorf("Non-zero input")
	}

	_, err := client.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret/verify",
			Body:   map[string]interface{}{"code": input},
		},
	)

	if err != nil {
		return err
	}
	n.TFACode = input
	return nil
}
