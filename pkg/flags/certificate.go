package flags

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/spf13/cobra"
)

// NewCertificateContents returns iterator which read a x509 certificate from file
func NewCertificateContents(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	supportsPipeline := HasValueFromPipeline(cmd, name)
	if cmd.Flags().Changed(name) {
		if path, err := cmd.Flags().GetString(name); err == nil && path != "" {

			iter := iterator.NewFuncIterator(func(i int64) (string, error) {
				r, err := os.ReadFile(path)

				if err != nil {
					return "", err
				}

				block, _ := pem.Decode(r)

				if block == nil || block.Type != "PUBLIC KEY" {
					return "", fmt.Errorf("failed to decode PEM block containing public key")
				}

				pub, err := x509.ParsePKIXPublicKey(block.Bytes)
				if err != nil {
					log.Fatal(err)
				}

				switch pub := pub.(type) {
				case *rsa.PublicKey:
					fmt.Println("pub is of type RSA:", pub)
				case *dsa.PublicKey:
					fmt.Println("pub is of type DSA:", pub)
				case *ecdsa.PublicKey:
					fmt.Println("pub is of type ECDSA:", pub)
				case ed25519.PublicKey:
					fmt.Println("pub is of type Ed25519:", pub)
				default:
					return "", fmt.Errorf("unknown type of public key")
				}

				return base64.StdEncoding.EncodeToString(block.Bytes), nil
			}, 1)

			return iter, nil
		}
	} else if supportsPipeline {
		return iterator.NewPipeIterator(cmd.InOrStdin())
	}
	return nil, fmt.Errorf("no input detected")
}
