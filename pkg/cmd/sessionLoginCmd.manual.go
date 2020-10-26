package cmd

import (
	"context"
	"fmt"
	"io"
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

type SessionHandler interface {
	Init()
	Login() error
	Verify() error
}

type LoginState int

func (l LoginState) String() string {
	return [...]string{"Unknown", "Authorized", "NotAuthorized", "TFASetup", "TFAConfirm", "Verify", "Abort"}[l]
}

const (
	LoginStateUnknown LoginState = iota
	LoginStateAuth
	LoginStateNoAuth
	LoginStateTFASetup
	LoginStateTFAConfirm
	LoginStateVerify
	LoginStateAbort
)

type loginHandler struct {
	TFACodeRequired bool
	Authorized      bool
	Err             error
	TFACode         string
	Cookies         []*http.Cookie
	C8Yclient       *c8y.Client
	LoginOptions    *c8y.TenantLoginOptions
	state           chan LoginState
	Attempts        int
	Writer          io.Writer
}

func NewLoginHandler(c *c8y.Client, w io.Writer) *loginHandler {
	h := &loginHandler{
		C8Yclient: c,
		Writer:    w,
	}
	h.state = make(chan LoginState, 1)
	return h
}

func (lh *loginHandler) do(op func() error) {
	// if lh.Err != nil {
	// 	return
	// }
	lh.Err = op()
}

func (lh *loginHandler) Clear() {
	lh.TFACode = ""
	lh.Authorized = false
	lh.C8Yclient.SetCookies([]*http.Cookie{})
	WriteAuth(viper.GetViper())
	lh.Err = nil
}

func (lh *loginHandler) Run() error {
	lh.Init()
	lh.state <- LoginStateUnknown

	for {
		c := <-lh.state

		Logger.Infof("Current State: %s\n", c.String())

		if c == LoginStateUnknown || c == LoginStateVerify {
			lh.Verify()
		} else if c == LoginStateTFAConfirm {
			lh.Login()
		} else if c == LoginStateNoAuth {
			lh.Clear()

			// login
			lh.Login()

		} else if c == LoginStateTFASetup {
			if err := lh.setupTFA(); err != nil {
				return fmt.Errorf("TFA Setup failed")
			}
			lh.state <- LoginStateTFAConfirm
		} else if c == LoginStateAbort {
			break
		} else if c == LoginStateAuth {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return lh.Err
}

func (lh *loginHandler) sortLoginOptions() {
	if lh.LoginOptions == nil {
		return
	}

	optionOrder := map[string]int{
		c8y.AuthMethodBasic:          2,
		c8y.AuthMethodOAuth2Internal: 1,
	}

	// sort login options
	sort.SliceStable(lh.LoginOptions.LoginOptions[:], func(i, j int) bool {
		iWeight := 100
		jWeight := 200

		if v, ok := optionOrder[lh.LoginOptions.LoginOptions[i].Type]; ok {
			iWeight = v
		}

		if v, ok := optionOrder[lh.LoginOptions.LoginOptions[j].Type]; ok {
			jWeight = v
		}
		return iWeight < jWeight
	})
}

func (lh *loginHandler) Init() {
	lh.do(func() error {
		loginOptions, _, err := client.Tenant.GetLoginOptions(context.Background())
		lh.LoginOptions = loginOptions
		lh.sortLoginOptions()
		return err
	})
}

func (lh *loginHandler) Login() {
	lh.do(func() error {
		if lh.LoginOptions == nil && len(lh.LoginOptions.LoginOptions) == 0 {
			lh.state <- LoginStateAbort
			return fmt.Errorf("No login options")
		}

		if lh.Authorized {
			lh.state <- LoginStateAuth
			return nil
		}

		for _, option := range lh.LoginOptions.LoginOptions {
			switch option.Type {
			case c8y.AuthMethodOAuth2Internal:

				if lh.TFACodeRequired && option.TFAStrategy == "TOTP" {
					if lh.TFACode == "" {
						prompt := promptui.Prompt{
							Stdin:   os.Stdin,
							Stdout:  os.Stderr,
							Default: lh.TFACode,
							Label:   "Enter Two-Factor code",
							Validate: func(input string) error {
								if len(strings.ReplaceAll(input, " ", "")) < 6 {
									return fmt.Errorf("Non-zero input")
								}

								ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
								defer cancel()

								lh.C8Yclient.TFACode = input

								if err := lh.C8Yclient.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
									Logger.Errorf("OAuth2 failed. %s", err)
									return err
								}
								lh.TFACode = input
								lh.Authorized = true
								return nil
							},
						}

						if v, err := prompt.Run(); err == nil {
							lh.TFACode = v
						} else {
							lh.state <- LoginStateAbort
							lh.Err = fmt.Errorf("User cancelled login")
							return nil
						}
					}
					lh.C8Yclient.TFACode = lh.TFACode
				}

				if !lh.Authorized {
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
					defer cancel()

					Logger.Debugf("Logging in using %s", c8y.AuthMethodOAuth2Internal)
					if err := client.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
						lh.Attempts++

						if v, ok := err.(*c8y.ErrorResponse); ok {
							Logger.Errorf("OAuth2 failed. %s", v.Message)
						} else {
							Logger.Errorf("OAuth2 failed. %s", err)
						}

						if strings.Contains(err.Error(), "There was a change in authentication strategy for your tenant or user account") {
							lh.state <- LoginStateNoAuth
							return nil
						}

						if lh.Attempts > 2 {
							lh.Err = fmt.Errorf("Max log attempts reached: %w", err)
							lh.state <- LoginStateAbort
							return nil
						}

						// trigger unknown to recheck if TFA is required or not
						lh.Clear()
						lh.state <- LoginStateUnknown
						return nil
					}
					lh.Authorized = true
					lh.state <- LoginStateAuth
					WriteAuth(viper.GetViper())
					return nil
				}
				WriteAuth(viper.GetViper())
				lh.state <- LoginStateVerify

			case c8y.AuthMethodBasic:
				// do nothing
			}
			break
		}
		return nil
	})
}

func (lh *loginHandler) Verify() {
	lh.do(func() error {
		_, resp, err := client.User.GetCurrentUser(context.Background())

		if resp != nil && resp.StatusCode == http.StatusUnauthorized {

			if v, ok := err.(*c8y.ErrorResponse); ok {

				if strings.Contains(v.Message, "TFA TOTP setup required") {
					lh.TFACodeRequired = true
					lh.state <- LoginStateTFASetup
				} else if strings.Contains(v.Message, "TFA TOTP code required") {
					Logger.Debug("TFA code is required. server response: %s", v.Message)
					lh.TFACodeRequired = true
					lh.state <- LoginStateNoAuth
				} else {
					lh.state <- LoginStateNoAuth
				}
				return nil
			}

			lh.state <- LoginStateNoAuth
		} else {
			lh.state <- LoginStateAuth
		}
		return err
	})
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
		RunE: ccmd.initSession,
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

func (n *sessionLoginCmd) initSession(cmd *cobra.Command, args []string) error {
	handler := NewLoginHandler(client, cmd.ErrOrStderr())
	return handler.Run()
}

func (lh *loginHandler) setupTFA() error {

	// Request TFA secret
	backupAuthMethod := lh.C8Yclient.AuthorizationMethod
	client.AuthorizationMethod = c8y.AuthMethodBasic
	resp, err := lh.C8Yclient.SendRequest(
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
			Writer:    lh.Writer,
			BlackChar: qrterminal.BLACK,
			WhiteChar: qrterminal.WHITE,
			QuietZone: 1,
		})

		fmt.Printf("\nTOTP Secret: %s\n\n", v.String())
		// n.cmd.Printf("\nTOTP Secret: %s\n\n", v.String())
	}

	// Verify TOTP by checking a code
	tfaCodePrompt := promptui.Prompt{
		Stdin:    os.Stdin,
		Stdout:   os.Stderr,
		Label:    "Enter Two-Factor code",
		Validate: lh.verifyTFASetupCode,
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

func (lh *loginHandler) verifyTFASetupCode(input string) error {
	if len(strings.ReplaceAll(input, " ", "")) < 6 {
		return fmt.Errorf("Non-zero input")
	}

	_, err := lh.C8Yclient.SendRequest(
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
	lh.TFACode = input
	return nil
}
