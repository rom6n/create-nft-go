package withdraw_user_ton

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type WithdrawUserTonRepository interface {
	Withdraw(ctx context.Context, userID int64, amount uint64, withdrawToAddress *address.Address, isTestnet bool) error
	WithdrawQueue()
}

type withdrawUserTonRepo struct {
	userRepo          user.UserRepository
	testnetLiteClient *liteclient.ConnectionPool
	mainnetLiteClient *liteclient.ConnectionPool
	testnetWallet     *wallet.Wallet
	mainnetWallet     *wallet.Wallet
	queueChannel      chan *WithdrawRequest
	timeout           time.Duration
}

type WithdrawUserTonCfg struct {
	UserRepo          user.UserRepository
	TestnetLiteClient *liteclient.ConnectionPool
	MainnetLiteClient *liteclient.ConnectionPool
	TestnetWallet     *wallet.Wallet
	MainnetWallet     *wallet.Wallet
	QueueChannel      chan *WithdrawRequest
	Timeout           time.Duration
}

func New(cfg WithdrawUserTonCfg) WithdrawUserTonRepository {
	return &withdrawUserTonRepo{
		userRepo:          cfg.UserRepo,
		testnetLiteClient: cfg.TestnetLiteClient,
		mainnetLiteClient: cfg.MainnetLiteClient,
		testnetWallet:     cfg.TestnetWallet,
		mainnetWallet:     cfg.MainnetWallet,
		queueChannel:      cfg.QueueChannel,
		timeout:           cfg.Timeout,
	}
}

type WithdrawRequest struct {
	Wallet            *wallet.Wallet
	Ctx               context.Context
	WithdrawToAddress *address.Address
	Amount            tlb.Coins
	UserUUID          uuid.UUID
	UserNanoTON       uint64
}

func (v *withdrawUserTonRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *withdrawUserTonRepo) Withdraw(ctx context.Context, userID int64, amount uint64, withdrawToAddress *address.Address, isTestnet bool) error {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	//client := v.testnetLiteClient
	w := v.testnetWallet
	if !isTestnet {
		w = v.mainnetWallet
		//client = v.mainnetLiteClient
	}

	//apiCtx := client.StickyContext(svcCtx)

	user, getErr := v.userRepo.GetUserByID(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("error getting user by ID: %w", getErr)
	}
	if user.NanoTon < amount {
		return fmt.Errorf("not enough balance")
	}

	updErr := v.userRepo.UpdateUserBalance(svcCtx, user.UUID, user.NanoTon-amount)
	if updErr != nil {
		return fmt.Errorf("error updating user's balance 2: %w", updErr)
	}

	go func() {
		v.queueChannel <- &WithdrawRequest{
			Wallet:            w,
			WithdrawToAddress: withdrawToAddress,
			Ctx:               context.Background(),
			UserUUID:          user.UUID,
			UserNanoTON:       user.NanoTon,
			Amount:            tlb.FromNanoTONU(amount),
		}
	}()

	return nil
}

func (v *withdrawUserTonRepo) WithdrawQueue() {
	log.Printf("Withdraw queue is running")
	for {
		select {
		case request := <-v.queueChannel:
			if transferErr := request.Wallet.Transfer(request.Ctx, request.WithdrawToAddress, request.Amount, "Thanks for using Build NFT tma"); transferErr != nil {
				updErr := v.userRepo.UpdateUserBalance(request.Ctx, request.UserUUID, request.UserNanoTON)
				if updErr != nil {
					log.Printf("error updating user's balance 2: %v. error withdrawing ton: %v", updErr, transferErr)
					continue
				}
				log.Printf("error withdrawing ton: %v", transferErr)
				continue
			}
			time.Sleep(10 * time.Second)
		default:
			continue
		}
	}
}
