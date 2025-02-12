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

// secretUpdateCmd represents the "secret update" command.
var secretUpdateCmd = &cobra.Command{
	Use:   "update [type]",
	Short: "Update an existing secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretType := args[0] // card, credentials, text, bin

		// Need secret id for update.
		secretID, err := cmd.Flags().GetString("id")
		if err != nil || secretID == "" {
			logging.Sugar.Fatal("Secret id (--id) must be provided")
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil || name == "" {
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
			pass, _ := cmd.Flags().GetString("password")
			rawData = fmt.Sprintf("name:%s;login:%s;password:%s", name, login, pass)
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
			logging.Sugar.Fatal("Encryption key (encryption_key) is not set in configuration")
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
		// Формируем запрос для обновления секрета.
		req := &proto.EditSecretRequest{
			Id: secretID, // Здесь передаем идентификатор секрета (как строку, если в proto оно строковое).
			Secret: &proto.Secret{
				Data: encryptedData,
				Meta: name,
			},
		}

		// Create context with token in metadata.
		md := metadata.Pairs("cookie", token)
		ctx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), md), 5*time.Second)
		defer cancel()

		_, err = client.EditSecret(ctx, req)
		if err != nil {
			logging.Sugar.Fatalf("Failed to update secret: %v", err)
		}

		fmt.Printf("Secret updated successfully (id: %s)\n", secretID)
	},
}

func init() {
	secretCmd.AddCommand(secretUpdateCmd)

	// Flags for all types.
	secretUpdateCmd.Flags().StringP("id", "i", "", "Secret identifier (id) to update")
	secretUpdateCmd.MarkFlagRequired("id")

	secretUpdateCmd.Flags().StringP("name", "n", "", "Name for the secret")
	secretUpdateCmd.MarkFlagRequired("name")

	// Type card.
	secretUpdateCmd.Flags().String("number", "", "Card number")
	secretUpdateCmd.Flags().String("date", "", "Card expiration date (MM/YY)")
	secretUpdateCmd.Flags().String("holder", "", "Card holder name")
	secretUpdateCmd.Flags().String("code", "", "Card security code")

	// Type credentials.
	secretUpdateCmd.Flags().String("login", "", "Login for credentials")
	secretUpdateCmd.Flags().String("password", "", "Password for credentials")

	// Type text.
	secretUpdateCmd.Flags().String("data", "", "Text data")

	// Type bin.
	secretUpdateCmd.Flags().StringP("file", "f", "", "File path for binary data")
}
