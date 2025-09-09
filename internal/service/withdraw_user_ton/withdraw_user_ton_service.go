package withdraw_user_ton

import (
	"context"
	"fmt"
	"time"

	"github.com/rom6n/create-nft-go/internal/domain/user"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type WithdrawUserTonRepository interface {
	Withdraw(ctx context.Context, userID int64, amount uint64, withdrawToAddress *address.Address, isTestnet bool) error
}

type withdrawUserTonRepo struct {
	userRepo          user.UserRepository
	testnetLiteClient *liteclient.ConnectionPool
	mainnetLiteClient *liteclient.ConnectionPool
	testnetWallet     *wallet.Wallet
	mainnetWallet     *wallet.Wallet
	timeout           time.Duration
}

type WithdrawUserTonCfg struct {
	UserRepo          user.UserRepository
	TestnetLiteClient *liteclient.ConnectionPool
	MainnetLiteClient *liteclient.ConnectionPool
	TestnetWallet     *wallet.Wallet
	MainnetWallet     *wallet.Wallet
	Timeout           time.Duration
}

func New(cfg WithdrawUserTonCfg) WithdrawUserTonRepository {
	return &withdrawUserTonRepo{
		userRepo:          cfg.UserRepo,
		testnetLiteClient: cfg.TestnetLiteClient,
		mainnetLiteClient: cfg.MainnetLiteClient,
		testnetWallet:     cfg.TestnetWallet,
		mainnetWallet:     cfg.MainnetWallet,
		timeout:           cfg.Timeout,
	}
}

func (v *withdrawUserTonRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *withdrawUserTonRepo) Withdraw(ctx context.Context, userID int64, amount uint64, withdrawToAddress *address.Address, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	client := v.testnetLiteClient
	w := v.testnetWallet
	if !isTestnet {
		w = v.mainnetWallet
		client = v.mainnetLiteClient
	}

	apiCtx := client.StickyContext(svcCtx)

	user, getErr := v.userRepo.GetUserByID(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("error getting user by ID: %w", getErr)
	}

	if user.NanoTon < amount {
		return fmt.Errorf("not enough balance")
	}

	updErr := v.userRepo.UpdateUserBalance(svcCtx, user.UUID, user.NanoTon - amount)
	if updErr != nil {
		return fmt.Errorf("error updating user's balance 2: %w", updErr)
	}

	if transferErr := w.Transfer(apiCtx, withdrawToAddress, tlb.FromNanoTONU(amount), "Thanks for using Build NFT tma"); transferErr != nil {
		updErr := v.userRepo.UpdateUserBalance(svcCtx, user.UUID, user.NanoTon)
		if updErr != nil {
			return fmt.Errorf("error updating user's balance 2: %w", updErr)
		}
		return fmt.Errorf("error withdrawing ton: %v", transferErr)
	}

	return nil
}
