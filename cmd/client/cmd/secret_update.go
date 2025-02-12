package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
		if err != nil {
			logging.Sugar.Fatal("Secret id (--id) must be provided")
		}

		note, _ := cmd.Flags().GetString("note")

		var rawData string
		switch secretType {
		case "card":
			number, _ := cmd.Flags().GetString("number")
			date, _ := cmd.Flags().GetString("date")
			holder, _ := cmd.Flags().GetString("holder")
			code, _ := cmd.Flags().GetString("code")
			rawData = fmt.Sprintf("number:%s;date:%s;holder:%s;code:%s", number, date, holder, code)
		case "credentials":
			login, _ := cmd.Flags().GetString("login")
			pass, _ := cmd.Flags().GetString("password")
			rawData = fmt.Sprintf("login:%s;password:%s", login, pass)
		case "text":
			data, _ := cmd.Flags().GetString("data")
			rawData = fmt.Sprintf("data:%s", data)
		case "bin":
			filePath, _ := cmd.Flags().GetString("file")
			content, err := os.ReadFile(filePath)
			if err != nil {
				logging.Sugar.Fatalf("Failed to read file: %v", err)
			}
			rawData = fmt.Sprintf("bin:%x", content)
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

		// Encrypt metadata using encryptionKey.
		encryptedMeta, err := encryption.EncryptWithKey(note, encryptionKey)
		if err != nil {
			logging.Sugar.Fatalf("Failed to encrypt metadata: %v", err)
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

		id, err := strconv.ParseInt(secretID, 10, 64)
		if err != nil {
			logging.Sugar.Fatalf("Failed to parse id: %v", err)
		}
		req := &proto.EditSecretRequest{
			Id: id,
			Secret: &proto.Secret{
				Data: encryptedData,
				Meta: encryptedMeta,
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

		fmt.Printf("Secret updated successfully (id: %d)\n", id)
	},
}

func init() {
	secretCmd.AddCommand(secretUpdateCmd)

	// Flags for all types.
	secretUpdateCmd.Flags().StringP("id", "i", "", "Secret identifier (id) to update")
	secretUpdateCmd.MarkFlagRequired("id")

	secretUpdateCmd.Flags().StringP("note", "n", "", "Optional note for the secret")

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
