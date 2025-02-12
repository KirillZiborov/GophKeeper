package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/encryption"
	"github.com/KirillZiborov/GophKeeper/internal/logging"
	"github.com/KirillZiborov/GophKeeper/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secret data",
	Long:  "Create, update and get your secrets from the GophKeeper service.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// secretUpdateCmd represents the "secret create" command.
var secretCreateCmd = &cobra.Command{
	Use:   "create [type]",
	Short: "Create a new secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretType := args[0] // secret type: card, credentials, text, bin.
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			logging.Sugar.Fatal("Name must be provided")
		}

		var rawData string
		switch secretType {
		case "card":
			number, _ := cmd.Flags().GetString("number")
			date, _ := cmd.Flags().GetString("date")
			holder, _ := cmd.Flags().GetString("holder")
			code, _ := cmd.Flags().GetString("code")
			rawData = fmt.Sprintf("name:%s;number:%s;date:%s;holder:%s;code:%s", name, number, date, holder, code)
		case "credentials":
			login, _ := cmd.Flags().GetString("login")
			password, _ := cmd.Flags().GetString("password")
			rawData = fmt.Sprintf("name:%s;login:%s;password:%s", name, login, password)
		case "text":
			data, _ := cmd.Flags().GetString("data")
			rawData = fmt.Sprintf("name:%s;data:%s", name, data)
		case "bin":
			filePath, _ := cmd.Flags().GetString("file")
			content, err := os.ReadFile(filePath)
			if err != nil {
				logging.Sugar.Fatalf("Failed to read file: %v", err)
			}
			rawData = fmt.Sprintf("name:%s;bin:%x", name, content)
		default:
			logging.Sugar.Fatalf("Unknown secret type: %s", secretType)
		}

		// Read encryption key from config.
		encryptionKey := viper.GetString("encryption_key")
		if encryptionKey == "" {
			logging.Sugar.Fatal("Encryption key is not set in configuration")
		}
		// Encrypt data using encryptionKey.
		encryptedData, err := encryption.EncryptWithKey(rawData, encryptionKey)
		if err != nil {
			logging.Sugar.Fatalf("Failed to encrypt data: %v", err)
		}

		// Read token from file (token.txt).
		tokenBytes, err := os.ReadFile("token.txt")
		if err != nil {
			logging.Sugar.Fatalf("Failed to read token file: %v", err)
		}
		token := strings.TrimSpace(string(tokenBytes))
		if token == "" {
			logging.Sugar.Fatal("Please login first: no token")
		}

		conn, err := grpc.NewClient(
			viper.GetString("grpc_address"),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logging.Sugar.Fatalf("Failed to connect gRPC server: %v", err)
		}
		defer conn.Close()

		client := proto.NewKeeperClient(conn)

		secretData := &proto.Secret{
			Data: encryptedData,
			Meta: name,
		}

		req := &proto.AddSecretRequest{
			Secret: secretData,
		}

		// Create context with token in metadata.
		md := metadata.Pairs("cookie", token)
		ctx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), md), 5*time.Second)
		defer cancel()

		resp, err := client.AddSecret(ctx, req)
		if err != nil {
			logging.Sugar.Fatalf("Failed to add secret: %v", err)
		}

		fmt.Printf("Secret created with id: %s\n", resp.Id)
	},
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretCreateCmd)

	// Flags for all types.
	secretCreateCmd.Flags().StringP("name", "n", "", "Name for the secret")
	secretCreateCmd.MarkFlagRequired("name")

	// Type card.
	secretCreateCmd.Flags().String("number", "", "Card number")
	secretCreateCmd.Flags().String("date", "", "Card expiration date (MM/YY)")
	secretCreateCmd.Flags().String("holder", "", "Card holder name")
	secretCreateCmd.Flags().String("code", "", "Card security code")

	// Type credentials.
	secretCreateCmd.Flags().String("login", "", "Login for credentials")
	secretCreateCmd.Flags().String("password", "", "Password for credentials")

	// Type text.
	secretCreateCmd.Flags().String("data", "", "Text data")

	// Type bin.
	secretCreateCmd.Flags().StringP("file", "f", "", "File path for binary data")
}
