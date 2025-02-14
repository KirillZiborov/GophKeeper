package grpcapi_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/app"
	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi"
	"github.com/KirillZiborov/GophKeeper/internal/models"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/KirillZiborov/GophKeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// Test case: Login request with wrong credentials.
func TestWrongCredentialsGRPC(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := app.KeeperService{
		Store: fakeStore,
	}

	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor()))
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&svc))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("gRPC server exited with error")
		}
	}()
	defer grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient(
		"bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewKeeperClient(conn)

	// Login request without previous registration.
	loginReq := &proto.LoginRequest{
		UserData: &proto.User{
			Username: "abcdef",
			Password: "abcdefgh",
		},
	}

	var loginHeader metadata.MD
	_, err = client.Login(ctx, loginReq, grpc.Header(&loginHeader))

	// Expect error.
	require.Error(t, err, "Expected error when login credentials are wrong")
	// Expect status code Unauthenticated.
	st, ok := status.FromError(err)
	code := st.Code()
	require.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, code, codes.Unauthenticated)
}

// Test case: Register request and Login request with the same credentials.
func TestLoginGRPC(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := app.KeeperService{
		Store: fakeStore,
	}

	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor()))
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&svc))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("gRPC server exited with error")
		}
	}()
	defer grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient(
		"bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewKeeperClient(conn)

	// Register request.
	regReq := &proto.RegisterRequest{
		UserData: &proto.User{
			Username: "testuser",
			Password: "testpassword",
		},
	}

	var regHeader metadata.MD
	_, err = client.Register(ctx, regReq, grpc.Header(&regHeader))
	require.NoError(t, err)

	tokens := regHeader.Get("token")
	require.NotEmpty(t, tokens, "Expected token in header after registration")

	// Login request with the same credentials.
	loginReq := &proto.LoginRequest{
		UserData: &proto.User{
			Username: "testuser",
			Password: "testpassword",
		},
	}

	var loginHeader metadata.MD
	_, err = client.Login(ctx, loginReq, grpc.Header(&loginHeader))
	// Expect success.
	require.NoError(t, err)

	tokens = loginHeader.Get("token")
	require.NotEmpty(t, tokens, "Expected token in header after login")
}

// Test case: Register request and Login request with the wrong password.
func TestLoginWrongPassGRPC(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := app.KeeperService{
		Store: fakeStore,
	}

	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor()))
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&svc))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("gRPC server exited with error")
		}
	}()
	defer grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient(
		"bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewKeeperClient(conn)

	// Register request.
	regReq := &proto.RegisterRequest{
		UserData: &proto.User{
			Username: "testuser",
			Password: "correct",
		},
	}

	var regHeader metadata.MD
	_, err = client.Register(ctx, regReq, grpc.Header(&regHeader))
	// Expect success.
	require.NoError(t, err)

	tokens := regHeader.Get("token")
	require.NotEmpty(t, tokens, "Expected token in header after registration")

	// Login request with the wrong password.
	loginReq := &proto.LoginRequest{
		UserData: &proto.User{
			Username: "testuser",
			Password: "wrong",
		},
	}

	var loginHeader metadata.MD
	_, err = client.Login(ctx, loginReq, grpc.Header(&loginHeader))

	// Expect error.
	require.Error(t, err, "Expected error when login credentials are wrong")
	st, ok := status.FromError(err)
	// Expect status code Unauthenticated.
	code := st.Code()
	require.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, code, codes.Unauthenticated)
}

// Test GophKeeper server main functionalities.
func TestSecretCRUDGRPC(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("test-secret", "2h")

	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor()))
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&svc))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("gRPC server exited with error")
		}
	}()
	defer grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient(
		"bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewKeeperClient(conn)

	// Register first.
	regReq := &proto.RegisterRequest{
		UserData: &proto.User{
			Username: "testuser",
			Password: "testpassword",
		},
	}

	var regHeader metadata.MD
	_, err = client.Register(ctx, regReq, grpc.Header(&regHeader))
	// Expect success.
	require.NoError(t, err)

	// Expect success extracting token.
	tokens := regHeader.Get("token")
	require.NotEmpty(t, tokens, "Expected token in header after registration")
	require.NotEmpty(t, tokens[0], "Expected token in header after registration")

	// Put the token to the context as a real client.
	md := metadata.Pairs("token", tokens[0])
	authCtx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), md), 5*time.Second)
	defer cancel()

	// AddSecret request with token.
	secretData := &proto.Secret{
		Data: "encryptedData",
		Meta: "encryptedMeta",
	}

	addReq := &proto.AddSecretRequest{
		Secret: secretData,
	}

	addResp, err := client.AddSecret(authCtx, addReq)
	// Expect success.
	require.NoError(t, err)
	// Expect success getting generated secret ID.
	require.NotEmpty(t, addResp.Id, "Expected non-empty secret id from CreateData")
	secretID := addResp.Id

	// UpdateSecret request.
	updateReq := &proto.EditSecretRequest{
		Id: secretID,
		Secret: &proto.Secret{
			Data: "updatedEncryptedData",
			Meta: "updated note",
		},
	}
	_, err = client.EditSecret(authCtx, updateReq)
	// Expect success.
	require.NoError(t, err, "Expected UpdateSecret to succeed")

	// GetSecret request.
	getReq := &proto.GetSecretRequest{}
	getResp, err := client.GetSecret(authCtx, getReq)
	// Expect success.
	require.NoError(t, err, "Expected GetCredentials to succeed")
	require.GreaterOrEqual(t, len(getResp.Secret), 1, "Expected at least one secret")

	// Check that secret with id updated successfully.
	var found *proto.CountedSecret
	for _, cred := range getResp.Secret {
		if cred.Id == secretID {
			found = cred
			break
		}
	}

	require.NotNil(t, found, "Secret with given id not found")
	assert.Equal(t, "updated note", found.Secret.Meta, "Secret meta should be updated")
	assert.Equal(t, "updatedEncryptedData", found.Secret.Data, "Secret data should be updated")
}

// Test case: UpdateSecret request on other's user secret.
func TestUpdateNotMySecret(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("test-secret", "2h")

	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor()))
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&svc))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("gRPC server exited with error")
		}
	}()
	defer grpcServer.GracefulStop()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient(
		"bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewKeeperClient(conn)

	// Register user1 and user2.
	user1 := &models.User{
		ID:       "user1",
		Username: "user1",
		Password: "password",
	}
	user2 := &models.User{
		ID:       "user2",
		Username: "user2",
		Password: "password",
	}

	err = fakeStore.RegisterUser(user1)
	require.NoError(t, err)
	err = fakeStore.RegisterUser(user2)
	require.NoError(t, err)

	// Generate tokens for user1 and user2.
	tokenUser1, err := auth.GenerateToken(user1.ID)
	require.NoError(t, err)
	tokenUser2, err := auth.GenerateToken(user2.ID)
	require.NoError(t, err)

	mdUser1 := metadata.Pairs("token", tokenUser1)
	ctxUser1 := metadata.NewOutgoingContext(context.Background(), mdUser1)

	// user1 creates secret.
	secretData := &proto.Secret{
		Data: "encryptedData",
		Meta: "encryptedMeta",
	}

	addReq := &proto.AddSecretRequest{
		Secret: secretData,
	}

	addResp, err := client.AddSecret(ctxUser1, addReq)
	// Expect success.
	require.NoError(t, err)
	require.NotEmpty(t, addResp.Id, "Expected non-empty secret id from AddSecret")
	secretID := addResp.Id

	// user2 tries to update user1 secret.
	mdUser2 := metadata.Pairs("token", tokenUser2)
	ctxUser2 := metadata.NewOutgoingContext(context.Background(), mdUser2)

	updateReq := &proto.EditSecretRequest{
		Id: secretID,
		Secret: &proto.Secret{
			Data: "updatedEncryptedData",
			Meta: "updated note",
		},
	}
	_, err = client.EditSecret(ctxUser2, updateReq)
	// Expect error.
	require.Error(t, err, "Expected error when updating secret not belonging to the user")
	st, ok := status.FromError(err)
	code := st.Code()
	require.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, code, codes.Internal)
}
