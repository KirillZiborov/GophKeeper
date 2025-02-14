package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/logging"
	"github.com/KirillZiborov/GophKeeper/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Signs in a user in the GophKeeper service",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.NewClient(
			viper.GetString("grpc_address"),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logging.Sugar.Fatalf("Failed to connect gRPC server: %v", err)
		}
		defer conn.Close()

		client := proto.NewKeeperClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		username, err := cmd.Flags().GetString("username")
		if err != nil {
			logging.Sugar.Fatalw("Failed to read username")
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			logging.Sugar.Fatalw("Failed to read password")
		}

		userData := &proto.User{
			Username: username,
			Password: password,
		}

		var headerMD metadata.MD

		_, err = client.Login(ctx, &proto.LoginRequest{
			UserData: userData,
		}, grpc.Header(&headerMD))

		if err != nil {
			if err != nil && strings.Contains(err.Error(), "not found") {
				fmt.Println("Wrong username or password")
				return
			}
			logging.Sugar.Fatalf("Login failed: %v", err)
		}

		// Extract token from response header.
		tokens := headerMD.Get("token")
		if len(tokens) == 0 {
			logging.Sugar.Error("Token not found in response header")
		} else {
			token := tokens[0]
			fmt.Printf("Access Token: %s\n", token)

			if err := tokenStorage.Save(token); err != nil {
				logging.Sugar.Fatalw("Failed to store access token")
			}
		}

		fmt.Println("Login successfully")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	registerCmd.Flags().StringP("username", "u", "", "User Email")
	if err := registerCmd.MarkFlagRequired("username"); err != nil {
		logging.Sugar.Error(err)
	}

	registerCmd.Flags().StringP("password", "p", "", "User password")
	if err := registerCmd.MarkFlagRequired("password"); err != nil {
		logging.Sugar.Error(err)
	}
}
