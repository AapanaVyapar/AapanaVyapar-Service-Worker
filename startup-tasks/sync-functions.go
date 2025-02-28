package services

import (
	"aapanavyapar-service-worker/helpers"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

/*
	In Below Function We Only Care For Context Error And This Is Because If Database Failed Then Message Should Not Get Acknowledge
	Else If There Is Some User Not Exist Or Other Operation Related Error We No Need To Care For That Because That Is Related To User Input Process.
	This Will Allow Us To Make Stream Empty From Bad Messages.
*/
func (dataSource *DataSources) PerformAddToFavorite(ctx context.Context, val *redis.XMessage) error {

	uId := val.Values["uId"].(string)
	prodId := val.Values["prodId"].(string)
	operation := val.Values["operation"].(string)

	likes, _ := strconv.ParseUint(val.Values["likes"].(string), 10, 64)

	fmt.Println("FAV : ", uId)
	fmt.Println("FAV : ", prodId)
	fmt.Println("FAV : ", operation)

	productId, err := primitive.ObjectIDFromHex(prodId)
	if err != nil {
		dataSource.Cash.AckFavToStream(ctx, val)
		dataSource.Cash.DelFromFavStream(ctx, val)
		return nil
	}

	dataContext, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute))

	if operation == "+" {
		errData := dataSource.Data.AddToFavoritesUserData(dataContext, uId, productId)
		fmt.Println("FAV : ", errData)
		if errData != nil && helpers.ContextError(dataContext) != nil {
			cancel()
			return errData
		}

		errData = dataSource.Data.IncreaseLikesInProductData(dataContext, productId)
		fmt.Println("FAV : ", errData)
		if errData != nil && helpers.ContextError(dataContext) != nil {
			cancel()
			return errData
		}

		if errData == nil {
			likes += 1

			err = dataSource.Cash.UpdateLikeOfProduct(ctx, prodId, likes)
			fmt.Println("FAV : ", err)
			if err != nil && helpers.ContextError(dataContext) != nil {
				cancel()
				return err
			}
		}

	} else {
		errData := dataSource.Data.DelFromFavoritesUserData(dataContext, uId, productId)
		if errData != nil && helpers.ContextError(dataContext) != nil {
			fmt.Println("FAV : ", errData)
			cancel()
			return errData
		}

		errData = dataSource.Data.DecreaseLikesInProductData(dataContext, productId)
		if errData != nil && helpers.ContextError(dataContext) != nil {
			fmt.Println("FAV : ", errData)
			cancel()
			return errData
		}

		if errData == nil {
			if likes != 0 {
				likes -= 1

				err = dataSource.Cash.UpdateLikeOfProduct(ctx, prodId, likes)
				fmt.Println("FAV : ", err)
				if err != nil && helpers.ContextError(dataContext) != nil {
					cancel()
					return err
				}
			}
		}
	}

	fmt.Println("FAV : Done Acknowledge")
	dataSource.Cash.AckFavToStream(ctx, val)
	dataSource.Cash.DelFromFavStream(ctx, val)
	fmt.Println("FAV :  ", time.Now().UTC())
	cancel()
	return nil
}

/*
	In Below Function We Only Care For Context Error And This Is Because If Database Failed Then Message Should Not Get Acknowledge
	Else If There Is Some User Not Exist Or Other Operation Related Error We No Need To Care For That Because That Is Related To User Input Process.
	This Will Allow Us To Make Stream Empty From Bad Messages.
*/
func (dataSource *DataSources) PerformAddToCart(ctx context.Context, val *redis.XMessage) error {

	uId := val.Values["uId"].(string)
	prodId := val.Values["prodId"].(string)
	operation := val.Values["operation"].(string)

	fmt.Println("CART : ", uId)
	fmt.Println("CART : ", prodId)
	fmt.Println("CART : ", operation)

	productId, err := primitive.ObjectIDFromHex(prodId)
	if err != nil {
		dataSource.Cash.AckCartToStream(ctx, val)
		dataSource.Cash.DelFromCartStream(ctx, val)
		return nil
	}

	dataContext, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute))

	if operation == "+" {
		err := dataSource.Data.AddToCartUserData(dataContext, uId, productId)
		fmt.Println("CART : ", err)
		if err != nil && helpers.ContextError(dataContext) != nil {
			cancel()
			return err
		}

	} else {
		err := dataSource.Data.DelFromCartUserData(dataContext, uId, productId)
		if err != nil && helpers.ContextError(dataContext) != nil {
			fmt.Println("CART : ", err)
			cancel()
			return err
		}

	}

	fmt.Println("CART :  Done Acknowledge")
	dataSource.Cash.AckCartToStream(ctx, val)
	dataSource.Cash.DelFromCartStream(ctx, val)
	fmt.Println("CART :  ", time.Now().UTC())
	cancel()
	return nil
}
