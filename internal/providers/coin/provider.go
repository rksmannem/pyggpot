package coin_provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/aspiration-labs/pyggpot/internal/models"
	coin_service "github.com/aspiration-labs/pyggpot/rpc/go/coin"
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

	// get potID and count information from request
	potID := request.PotId
	count := request.Count

	// first query db for all the coins in particular pot, selected by potID
	coinsInPot, err := models.CoinsInPotsByPot_id(s.DB, int(potID))

	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	// coinCounter is a map where key is the denomination and value is the coin count at that denomination
	coinCounter := make(map[int32]int32)
	// totalCoins keeps track of total coins in the pot. Can probably be retrieved from SQL
	var totalCoins int32
	// dbModelCoins is a map where key is the denomination and value is a point to models.Coin
	dbModelCoins := make(map[int32]*models.Coin)

	for _, coin := range coinsInPot {
		// as we iterate through coinsInPot, populate coinCounter and increment totalCoins
		coinCounter[coin.Denomination] = coin.CoinCount
		totalCoins += coin.CoinCount
		// retrieve the coin model, selected by coin.ID, this step is probably not necessary if
		// we re-write models.CoinsInPotsByPot_id
		modelCoin, err := models.CoinByID(s.DB, coin.ID)
		if err != nil {
			return nil, twirp.InternalError(err.Error())
		}
		// populate dbModelCoins
		dbModelCoins[modelCoin.Denomination] = modelCoin
	}

	// begin shaking the piggy bank
	for i := 0; i < int(count); i++ {
		// simulate chance using rng.
		chance := rand.Float64()
		// iterate through all coins in coinCounter
		for denom, count := range coinCounter {
			// prob is the % distribution of a coin in the pot
			prob := float64(count) / float64(totalCoins)
			// we decrement chance by prob, if chance decrements below 0, we have
			// randomly shaken the piggy bank such that that denomination falls out
			chance -= prob
			if chance < 0 {
				coinCounter[denom]--
				totalCoins--
				// if last coin is removed, delete it from the map
				if coinCounter[denom] == 0 {
					delete(coinCounter, denom)
				}
				break
			}
		}
	}

	// begin process to update DB
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	// if the coin no longer exists in the piggy bank, delete it from db
	for denom, coin := range dbModelCoins {
		if _, ok := coinCounter[denom]; !ok {
			err = coin.Delete(tx)
		}
	}

	// declare response slice
	var res []*coin_service.Coins

	// iterate through the coins in coinCounter
	for denom, count := range coinCounter {

		coin := dbModelCoins[denom]
		// update the CoinCount property on the coin model
		coin.CoinCount = count

		// save coin
		err = coin.Save(tx)
		if err != nil {
			return nil, twirp.InternalError(err.Error())
		}

		// add a ptr to a coin_service.Coins struct to the res slice
		res = append(res, &coin_service.Coins{
			Kind:  coin_service.Coins_Kind(denom),
			Count: count,
		})
	}

	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	return &coin_service.CoinsListResponse{
		Coins: res,
	}, nil
}
