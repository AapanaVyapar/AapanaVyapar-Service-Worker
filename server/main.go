package main

import (
	services "aapanavyapar-service-worker/startup-tasks"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SyncWithAddToFavorite(ctx context.Context, dataBase *services.DataSources, checkBackLog bool, group *sync.WaitGroup) {
	lastId := "0"
	defer group.Done()

	startFav = time.Now().UTC()

	for {
		myKeyId := ">" //For Undelivered Ids So that Each Consumer Get Unique Id.
		if checkBackLog {
			myKeyId = lastId

		}

		readGroup := dataBase.Cash.ReadFromFavStream(ctx, 10, 200, myKeyId)
		if readGroup.Err() != nil {
			if strings.Contains(readGroup.Err().Error(), "timeout") {
				fmt.Println("FAV : TimeOUT")
				continue

			}
			if readGroup.Err() == redis.Nil {
				fmt.Println("FAV : No Data Available")
				continue

			}
			panic(readGroup.Err())

		}

		data, err := readGroup.Result()
		if err != nil {
			panic(err)

		}

		if len(data[0].Messages) == 0 {
			checkBackLog = false
			fmt.Println("FAV : Started Checking For New Messages ..!!")
			continue

		}

		var val redis.XMessage
		fmt.Println("\n\n\n NEW : ", lastId)
		for _, val = range data[0].Messages {
			err = dataBase.PerformAddToFavorite(ctx, &val)
			if err != nil {
				fmt.Println("FAV : Context Error Please Check For Data Base Connectivity, Network Error Or Any Other Dependency Problem")
				checkBackLog = true
				val.ID = "0"
				break

			}

		}
		lastId = val.ID
	}
}

func SyncWithAddToCart(ctx context.Context, dataBase *services.DataSources, checkBackLog bool, group *sync.WaitGroup) {
	lastId := "0"
	defer group.Done()
	startCart = time.Now().UTC()

	for {
		myKeyId := ">" //For Undelivered Ids So that Each Consumer Get Unique Id.
		if checkBackLog {
			myKeyId = lastId

		}

		readGroup := dataBase.Cash.ReadFromCartStream(ctx, 10, 2000, myKeyId)
		if readGroup.Err() != nil {
			if strings.Contains(readGroup.Err().Error(), "timeout") {
				fmt.Println("CART : TimeOUT")
				continue

			}

			if readGroup.Err() == redis.Nil {
				fmt.Println("CART : No Data Available")
				continue

			}

			fmt.Println(readGroup)
			panic(readGroup.Err())
		}

		data, err := readGroup.Result()
		if err != nil {
			panic(err)

		}

		if len(data[0].Messages) == 0 {
			checkBackLog = false
			fmt.Println("Started Checking For New Messages ..!!")
			continue

		}

		var val redis.XMessage
		fmt.Println("\n\n\n NEW : ", lastId)
		for _, val = range data[0].Messages {
			err = dataBase.PerformAddToCart(ctx, &val)
			if err != nil {
				fmt.Println("Context Error Please Check For Data Base Connectivity, Network Error Or Any Other Dependency Problem")
				checkBackLog = true
				val.ID = "0"
				break

			}

		}
		lastId = val.ID
	}
}

var startCart = time.Now()
var startFav = time.Now()

func main() {

	checkBackLog, err := strconv.ParseBool(os.Getenv("REDIS_STREAM_CHECK_BACKLOG"))
	if err != nil {
		panic(err)
	}

	dataBaseFav := services.NewDataSource()
	dataBaseCart := services.NewDataSource()

	defer dataBaseFav.Data.Data.Disconnect(context.TODO())
	defer dataBaseCart.Data.Data.Disconnect(context.TODO())

	err = dataBaseFav.InitFavStream(context.TODO())
	if err != nil {
		panic(err)
	}

	err = dataBaseCart.InitCartStream(context.TODO())
	if err != nil {
		panic(err)
	}

	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			dataBaseFav.Data.Data.Disconnect(context.TODO())
			dataBaseCart.Data.Data.Disconnect(context.TODO())
			fmt.Println("\n\n\nStart Cart : ", startFav)
			fmt.Println("\n\n\nStart Fav : ", startCart)
			break
		}
	}()

	go SyncWithAddToCart(context.TODO(), dataBaseCart, checkBackLog, &waitGroup)
	go SyncWithAddToFavorite(context.TODO(), dataBaseFav, checkBackLog, &waitGroup)

	waitGroup.Wait()

}

//Different Connections
// FAV :   2021-04-10 15:08:28.211232546 +0000 UTC
// Start Cart :  2021-04-10 15:06:30.480562923 +0000 UTC
// Diff = 2

//Single Connection
//CART :   2021-04-10 15:15:36.040009232 +0000 UTC
//Start Cart :  2021-04-10 15:13:39.261079629 +0000 UTC
// Diff = 2

//Single Connection On Shard
//CART :   2021-04-10 15:24:24.531923312 +0000 UTC
//Start Cart :  2021-04-10 15:20:38.035263452 +0000 UTC
// Diff = 4

//Different Connection On Shard
//CART :   2021-04-10 15:34:45.318704813 +0000 UTC
//Start Cart :  2021-04-10 15:30:10.169448625 +0000 UTC
// Diff = 4

//Different Connection On Shard
//CART :   2021-04-10 17:51:48.696856556 +0000 UTC
//Start Cart :  2021-04-10 17:46:10.60028291 +0000 UTC
//Diff = 5
