package security

import (
	"context"

	pb "github.com/DimaJoyti/go-coffee/web3-wallet-backend/api/proto/security"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler handles gRPC requests for the security service
type GRPCHandler struct {
	pb.UnimplementedSecurityServiceServer
	service *Service
	logger  *logger.Logger
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(service *Service, logger *logger.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger.Named("security-grpc-handler"),
	}
}

// GenerateKeyPair generates a new key pair
func (h *GRPCHandler) GenerateKeyPair(ctx context.Context, req *pb.GenerateKeyPairRequest) (*pb.GenerateKeyPairResponse, error) {
	// Convert request
	generateReq := &models.GenerateKeyPairRequest{
		Chain: models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.GenerateKeyPair(ctx, generateReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to generate key pair: %v", err)
	}

	// Convert response
	return &pb.GenerateKeyPairResponse{
		PrivateKey: resp.PrivateKey,
		PublicKey:  resp.PublicKey,
		Address:    resp.Address,
	}, nil
}

// EncryptPrivateKey encrypts a private key
func (h *GRPCHandler) EncryptPrivateKey(ctx context.Context, req *pb.EncryptPrivateKeyRequest) (*pb.EncryptPrivateKeyResponse, error) {
	// Convert request
	encryptReq := &models.EncryptPrivateKeyRequest{
		PrivateKey: req.PrivateKey,
		Passphrase: req.Passphrase,
	}

	// Call service
	resp, err := h.service.EncryptPrivateKey(ctx, encryptReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to encrypt private key: %v", err)
	}

	// Convert response
	return &pb.EncryptPrivateKeyResponse{
		EncryptedKey: resp.EncryptedKey,
	}, nil
}

// DecryptPrivateKey decrypts a private key
func (h *GRPCHandler) DecryptPrivateKey(ctx context.Context, req *pb.DecryptPrivateKeyRequest) (*pb.DecryptPrivateKeyResponse, error) {
	// Convert request
	decryptReq := &models.DecryptPrivateKeyRequest{
		EncryptedKey: req.EncryptedKey,
		Passphrase:   req.Passphrase,
	}

	// Call service
	resp, err := h.service.DecryptPrivateKey(ctx, decryptReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to decrypt private key: %v", err)
	}

	// Convert response
	return &pb.DecryptPrivateKeyResponse{
		PrivateKey: resp.PrivateKey,
	}, nil
}

// GenerateJWT generates a JWT token
func (h *GRPCHandler) GenerateJWT(ctx context.Context, req *pb.GenerateJWTRequest) (*pb.GenerateJWTResponse, error) {
	// Convert request
	generateReq := &models.GenerateJWTRequest{
		UserID: req.UserId,
		Email:  req.Email,
		Role:   req.Role,
	}

	// Call service
	resp, err := h.service.GenerateJWT(ctx, generateReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to generate JWT token: %v", err)
	}

	// Convert response
	return &pb.GenerateJWTResponse{
		Token:        resp.Token,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}

// VerifyJWT verifies a JWT token
func (h *GRPCHandler) VerifyJWT(ctx context.Context, req *pb.VerifyJWTRequest) (*pb.VerifyJWTResponse, error) {
	// Convert request
	verifyReq := &models.VerifyJWTRequest{
		Token: req.Token,
	}

	// Call service
	resp, err := h.service.VerifyJWT(ctx, verifyReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to verify JWT token: %v", err)
	}

	// Convert response
	return &pb.VerifyJWTResponse{
		Valid:     resp.Valid,
		UserId:    resp.UserID,
		Email:     resp.Email,
		Role:      resp.Role,
		ExpiresAt: resp.ExpiresAt,
	}, nil
}

// GenerateMnemonic generates a mnemonic phrase
func (h *GRPCHandler) GenerateMnemonic(ctx context.Context, req *pb.GenerateMnemonicRequest) (*pb.GenerateMnemonicResponse, error) {
	// Convert request
	generateReq := &models.GenerateMnemonicRequest{
		Strength: int(req.Strength),
	}

	// Call service
	resp, err := h.service.GenerateMnemonic(ctx, generateReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to generate mnemonic: %v", err)
	}

	// Convert response
	return &pb.GenerateMnemonicResponse{
		Mnemonic: resp.Mnemonic,
	}, nil
}

// ValidateMnemonic validates a mnemonic phrase
func (h *GRPCHandler) ValidateMnemonic(ctx context.Context, req *pb.ValidateMnemonicRequest) (*pb.ValidateMnemonicResponse, error) {
	// Convert request
	validateReq := &models.ValidateMnemonicRequest{
		Mnemonic: req.Mnemonic,
	}

	// Call service
	resp, err := h.service.ValidateMnemonic(ctx, validateReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to validate mnemonic: %v", err)
	}

	// Convert response
	return &pb.ValidateMnemonicResponse{
		Valid: resp.Valid,
	}, nil
}

// MnemonicToPrivateKey converts a mnemonic to a private key
func (h *GRPCHandler) MnemonicToPrivateKey(ctx context.Context, req *pb.MnemonicToPrivateKeyRequest) (*pb.MnemonicToPrivateKeyResponse, error) {
	// Convert request
	convertReq := &models.MnemonicToPrivateKeyRequest{
		Mnemonic: req.Mnemonic,
		Path:     req.Path,
	}

	// Call service
	resp, err := h.service.MnemonicToPrivateKey(ctx, convertReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to convert mnemonic to private key: %v", err)
	}

	// Convert response
	return &pb.MnemonicToPrivateKeyResponse{
		PrivateKey: resp.PrivateKey,
		PublicKey:  resp.PublicKey,
		Address:    resp.Address,
	}, nil
}
