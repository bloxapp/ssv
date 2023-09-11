package cli

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/utils/rsaencryption"
	"github.com/spf13/cobra"
	"github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"go.uber.org/zap"
)

var generateOperatorKeysCmd = &cobra.Command{
	Use:   "generate-operator-keys",
	Short: "generates ssv operator keys",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logging.SetGlobalLogger("debug", "capital", "console", nil); err != nil {
			log.Fatal(err)
		}
		logger := zap.L().Named(logging.NameExportKeys)
		passwordFilePath, _ := cmd.Flags().GetString("password-file")
		privateKeyFilePath, _ := cmd.Flags().GetString("operator-key-file")

		pk, sk, err := rsaencryption.GenerateKeys()
		if err != nil {
			logger.Fatal("Failed to generate keys", zap.Error(err))
		}

		if privateKeyFilePath != "" {
			sk, pk, err = readPrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				logger.Fatal("Failed to read private key from file", zap.Error(err))
			}
		}

		if err := logging.SetGlobalLogger("debug", "capital", "console", nil); err != nil {
			logger.Fatal("", zap.Error(err))
		}

		if passwordFilePath != "" {
			encryptAndSavePrivateKey(sk, passwordFilePath, logger)
		} else {
			logger.Info("generated public key (base64)", zap.String("pk", base64.StdEncoding.EncodeToString(pk)))
			logger.Info("generated private key (base64)", zap.String("sk", base64.StdEncoding.EncodeToString(sk)))
		}
	},
}

func readPrivateKeyFromFile(filePath string) ([]byte, []byte, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := os.ReadFile(absPath)
	if err != nil {
		return nil, nil, err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(string(keyBytes))
	if err != nil {
		return nil, nil, err
	}

	keyPem, _ := pem.Decode(decodedBytes)
	rsaKey, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return nil, nil, err
	}

	skPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	operatorPublicKey, err := rsaencryption.ExtractPublicKey(rsaKey)
	if err != nil {
		return nil, nil, err
	}
	pk, err := base64.StdEncoding.DecodeString(operatorPublicKey)
	if err != nil {
		return nil, nil, err
	}
	return skPem, pk, nil
}

func encryptAndSavePrivateKey(sk []byte, passwordFilePath string, logger *zap.Logger) {
	absPath, err := filepath.Abs(passwordFilePath)
	if err != nil {
		logger.Fatal("Failed to read absolute path of password file", zap.Error(err))
	}

	passwordBytes, err := os.ReadFile(absPath)
	if err != nil {
		logger.Fatal("Failed to read password file", zap.Error(err))
	}

	encryptionPassword := string(passwordBytes)
	encryptedData, err := keystorev4.New().Encrypt(sk, encryptionPassword)
	if err != nil {
		logger.Fatal("Failed to encrypt private key", zap.Error(err))
	}

	encryptedJSON, err := json.Marshal(encryptedData)
	if err != nil {
		logger.Fatal("Failed to marshal encrypted data to JSON", zap.Error(err))
	}

	err = os.WriteFile("encrypted_private_key.json", encryptedJSON, 0600)
	if err != nil {
		logger.Fatal("Failed to write encrypted private key to file", zap.Error(err))
	}

	logger.Info("private key encrypted and stored in encrypted_private_key.json")
}

func init() {
	generateOperatorKeysCmd.Flags().StringP("password-file", "p", "", "File path to the password used to encrypt the private key")
	generateOperatorKeysCmd.Flags().StringP("operator-key-file", "o", "", "File path to the operator private key")
	RootCmd.AddCommand(generateOperatorKeysCmd)
}
