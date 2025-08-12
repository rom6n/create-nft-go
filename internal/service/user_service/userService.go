package userservice

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	nftcollection "github.com/rom6n/create-nft-go/internal/domain/nft_collection"
	nftitem "github.com/rom6n/create-nft-go/internal/domain/nft_item"
	"github.com/rom6n/create-nft-go/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserServiceRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*user.User, error)
	GetUserNftCollections(ctx context.Context, userID int64) []nftcollection.NftCollection
	GetUserNftItems(ctx context.Context, userID int64) []nftitem.NftItem
}

type userServiceRepo struct {
	userRepo          user.UserRepository
	nftCollectionRepo nftcollection.NftCollectionRepository
	nftItemRepo       nftitem.NftItemRepository
	timeout           time.Duration
}

type UserServiceCfg struct {
	UserRepo          user.UserRepository
	NftCollectionRepo nftcollection.NftCollectionRepository
	NftItemRepo       nftitem.NftItemRepository
	Timeout           time.Duration
}

func New(cfg UserServiceCfg) UserServiceRepository {
	return &userServiceRepo{
		userRepo:          cfg.UserRepo,
		nftCollectionRepo: cfg.NftCollectionRepo,
		nftItemRepo:       cfg.NftItemRepo,
		timeout:           cfg.Timeout,
	}
}

func (v *userServiceRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *userServiceRepo) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundUser, dbErr := v.userRepo.GetUserByID(svcCtx, userID)
	if dbErr != nil {
		if dbErr == mongo.ErrNoDocuments {
			newUuid := uuid.New()
			user := user.NewUser(newUuid, userID, 1, "user", 0)
			createErr := v.userRepo.CreateUser(svcCtx, &user)
			return &user, createErr
		}
		return nil, dbErr
	}

	return foundUser, nil
}

func (v *userServiceRepo) GetUserNftCollections(ctx context.Context, userID int64) []nftcollection.NftCollection {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	user, userErr := v.userRepo.GetUserByID(svcCtx, userID)
	if userErr != nil {
		log.Warnf("error fetching user's data: %v", userErr)
		return nil
	}

	nftCollections, nftCollectionsErr := v.nftCollectionRepo.GetNftCollectionsByOwnerUuid(svcCtx, user.UUID)
	if nftCollectionsErr != nil {
		log.Warnf("error fetching user's nft collections: %v", nftCollectionsErr)
		return nil
	}

	return nftCollections
}

func (v *userServiceRepo) GetUserNftItems(ctx context.Context, userID int64) []nftitem.NftItem {
	svcCtx, cancel := v.getContext(ctx)
	defer cancel()

	user, userErr := v.userRepo.GetUserByID(svcCtx, userID)
	if userErr != nil {
		log.Warnf("error fetching user's data: %v", userErr)
		return nil
	}

	nftItems, nftItemsErr := v.nftItemRepo.GetNftItemsByOwnerUuid(svcCtx, user.UUID)
	if nftItemsErr != nil {
		log.Warnf("error fetching user's nft items: %v", nftItemsErr)
		return nil
	}

	return nftItems
}
