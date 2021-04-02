package coin_provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/aspiration-labs/pyggpot/internal/models"
	"github.com/aspiration-labs/pyggpot/rpc/go/coin"
	"github.com/twitchtv/twirp"
)

type coinServer struct {
	DB *sql.DB
}

func New(db *sql.DB) *coinServer {
	return &coinServer{
		DB: db,
	}
}

func (s *coinServer) AddCoins(ctx context.Context, request *coin_service.AddCoinsRequest) (*coin_service.CoinsListResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError(err.Error(), "")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}
	for _, coin := range request.Coins {
		fmt.Println(coin)
		newCoin := models.Coin{
			PotID:        request.PotId,
			Denomination: int32(coin.Kind),
			CoinCount:    coin.Count,
		}
		err = newCoin.Save(tx)
		if err != nil {
			return nil, twirp.InvalidArgumentError(err.Error(), "")
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	return &coin_service.CoinsListResponse{
		Coins: request.Coins,
	}, nil
}

func (s *coinServer) RemoveCoins(ctx context.Context, request *coin_service.RemoveCoinsRequest) (*coin_service.CoinsListResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError(err.Error(), "")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	// check pot with pot_id exists, if not return error
	potId := request.GetPotId()
	if _, err := models.PotByID(s.DB, potId); err != nil {
		return nil, twirp.NotFoundError(fmt.Sprintf("pot with id: %v not exist", potId))
	}

	// if pot with pot_id exists, verify that pot contains requested COINs count to remove
	coinsInPot, err := models.CoinsInPotsByPot_id(s.DB, int(potId))
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	totalPotCoins := func(coinsInPot []*models.CoinsInPot) int32 {
		total := int32(0)
		for _, coinGroup := range coinsInPot {
			if coinGroup != nil {
				total += coinGroup.CoinCount
			}
		}
		return total
	}(coinsInPot)

	numOfCoinsToRemove := request.GetCount()
	if numOfCoinsToRemove > totalPotCoins {
		return nil, twirp.InvalidArgumentError(fmt.Sprintf("can not remove: %v coins because pot contains only: %v coins",
			numOfCoinsToRemove, totalPotCoins), "")
	}

	// remove the coins with different denominations from the pot
	toRemove := make(map[int32]int32)
	rand.Seed(time.Now().UnixNano())
	for numOfCoinsToRemove > 0 {
		// get random Coins_Kind and mark it as deleted
		kinds := CoinKinds()
		randKind := kinds[rand.Intn(len(kinds))]

		// check if this kind of coins are exists in the pot
		if found, at := Contains(coinsInPot, randKind); found {
			if coinsInPot[at].CoinCount > 0 {
				coinsInPot[at].CoinCount = coinsInPot[at].CoinCount - 1
				// decrement the total coins to remove by one
				numOfCoinsToRemove = numOfCoinsToRemove - 1
				// counts the number of coins deleted of each kind
				toRemove[randKind] += 1
			}
		}
	}

	// now update the coins table with the remaining coins of each kind
	for _, coinInPot := range coinsInPot {
		coin, err := models.CoinByID(s.DB, coinInPot.ID)
		if err != nil {
			return nil, err
		}

		err = coin.Save(tx)
		if err != nil {
			return nil, twirp.InvalidArgumentError(err.Error(), "")
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	var removed []*coin_service.Coins
	for kind, count := range toRemove {
		// append the removed coins
		removed = append(removed, &coin_service.Coins{
			Kind:  coin_service.Coins_Kind(kind),
			Count: count,
		})
	}

	return &coin_service.CoinsListResponse{Coins: removed}, nil
}

func CoinKinds() []int32 {

	v := make([]int32, 0, len(coin_service.Coins_Kind_name))

	for key, _ := range coin_service.Coins_Kind_name {
		v = append(v, key)
	}
	return v
}

func Contains(coins []*models.CoinsInPot, k int32) (bool, int) {
	for pos, c := range coins {
		if c.Denomination == k {
			return true, pos
		}
	}
	return false, -1
}
