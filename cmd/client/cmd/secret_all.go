package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/logging"
	"github.com/KirillZiborov/GophKeeper/pkg/encryption"
	"github.com/KirillZiborov/GophKeeper/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// DecryptedSecret is a structure for outputing user's saved secrets.
type DecryptedSecret struct {
	Id   int64  `json:"id"`
	Data string `json:"data"`
	Meta string `json:"meta"`
}

// secretAllCmd represents the "secret all" command.
var secretAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all secrets for the authenticated user",
	Long:  "Retrieves and displays a list of all secret data belonging to the authenticated user.",
	Run: func(cmd *cobra.Command, args []string) {
		// Read token from file (token.txt).
		tokenBytes, err := os.ReadFile("token.txt")
		if err != nil {
			logging.Sugar.Fatalf("Failed to read token file: %v", err)
		}
		token := strings.TrimSpace(string(tokenBytes))
		if token == "" {
			logging.Sugar.Fatal("Token is empty; please login first")
		}

		conn, err := grpc.NewClient(
			viper.GetString("grpc_address"),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logging.Sugar.Fatalf("Failed to connect gRPC server: %v", err)
		}
		defer conn.Close()

		client := proto.NewKeeperClient(conn)

		// Create context with token in metadata.
		md := metadata.Pairs("cookie", token)
		ctx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), md), 5*time.Second)
		defer cancel()

		req := &proto.GetSecretRequest{}

		resp, err := client.GetSecret(ctx, req)
		if err != nil {
			logging.Sugar.Fatalf("Failed to get secrets: %v", err)
		}

		// Read encryption key from config.
		encryptionKey := viper.GetString("encryption_key")
		if encryptionKey == "" {
			logging.Sugar.Fatal("Encryption key (encryption_key) is not set in configuration")
		}

		var secrets []DecryptedSecret

		for _, cred := range resp.Secret {
			encryptedData := cred.Secret.Data
			encryptedMeta := cred.Secret.Meta
			data, err := encryption.DecryptWithKey(encryptedData, encryptionKey)
			if err != nil {
				logging.Sugar.Errorf("Failed to decrypt secret (id: %d): %v", cred.Id, err)
				continue
			}
			meta, err := encryption.DecryptWithKey(encryptedMeta, encryptionKey)
			if err != nil {
				logging.Sugar.Errorf("Failed to decrypt secret (id: %d): %v", cred.Id, err)
				continue
			}

			secret := DecryptedSecret{
				Id:   cred.Id,
				Data: data,
				Meta: meta,
			}
			secrets = append(secrets, secret)
		}

		// Translate result to JSON and output.
		output, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			logging.Sugar.Fatalf("Failed to marshal secrets: %v", err)
		}

		fmt.Println("Your secrets:")
		fmt.Println(string(output))
	},
}

func init() {
	secretCmd.AddCommand(secretAllCmd)
}
