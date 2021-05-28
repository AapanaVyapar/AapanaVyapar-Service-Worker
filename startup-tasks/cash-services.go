package services

import (
	"aapanavyapar-service-worker/configurations/redisdb"
	"context"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"time"
)

type RedisDataBase struct {
	Cash *redis.Client
}

func NewRedisClient() *RedisDataBase {
	return &RedisDataBase{Cash: redisdb.InitRedis()}
}

func (dataBase *RedisDataBase) CreateFavStream(ctx context.Context) error {
	return dataBase.Cash.XGroupCreateMkStream(ctx, os.Getenv("REDIS_STREAM_FAV_NAME"), os.Getenv("REDIS_STREAM_FAV_GROUP"), "$").Err()

}

func (dataBase *RedisDataBase) CreateCartStream(ctx context.Context) error {
	return dataBase.Cash.XGroupCreateMkStream(ctx, os.Getenv("REDIS_STREAM_CART_NAME"), os.Getenv("REDIS_STREAM_CART_GROUP"), "$").Err()

}

func (dataBase *RedisDataBase) AckFavToStream(ctx context.Context, val *redis.XMessage) {
	dataBase.Cash.XAck(ctx, os.Getenv("REDIS_STREAM_FAV_NAME"), os.Getenv("REDIS_STREAM_FAV_GROUP"), val.ID)

}

func (dataBase *RedisDataBase) DelFromFavStream(ctx context.Context, val *redis.XMessage) {
	dataBase.Cash.XDel(ctx, os.Getenv("REDIS_STREAM_FAV_NAME"), val.ID)

}

func (dataBase *RedisDataBase) ReadFromFavStream(ctx context.Context, count int64, timeout time.Duration, myKeyId string) *redis.XStreamSliceCmd {

	readGroup := dataBase.Cash.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    os.Getenv("REDIS_STREAM_FAV_GROUP"),
		Consumer: os.Getenv("REDIS_WORKER_NAME"),
		Streams:  []string{os.Getenv("REDIS_STREAM_FAV_NAME"), myKeyId},
		Count:    count,   // No Of Data To Retrieve
		Block:    timeout, //TimeOut
		NoAck:    false,
	})
	return readGroup

}

func (dataBase *RedisDataBase) AckCartToStream(ctx context.Context, val *redis.XMessage) {
	dataBase.Cash.XAck(ctx, os.Getenv("REDIS_STREAM_CART_NAME"), os.Getenv("REDIS_STREAM_CART_GROUP"), val.ID)

}

func (dataBase *RedisDataBase) DelFromCartStream(ctx context.Context, val *redis.XMessage) {
	dataBase.Cash.XDel(ctx, os.Getenv("REDIS_STREAM_CART_NAME"), val.ID)

}

func (dataBase *RedisDataBase) ReadFromCartStream(ctx context.Context, count int64, timeout time.Duration, myKeyId string) *redis.XStreamSliceCmd {

	readGroup := dataBase.Cash.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    os.Getenv("REDIS_STREAM_CART_GROUP"),
		Consumer: os.Getenv("REDIS_WORKER_NAME"),
		Streams:  []string{os.Getenv("REDIS_STREAM_CART_NAME"), myKeyId},
		Count:    count,   // No Of Data To Retrieve
		Block:    timeout, //TimeOut
		NoAck:    false,
	})
	return readGroup

}

func (dataBase *RedisDataBase) UpdateLikeOfProduct(ctx context.Context, productId string, likes uint64) error {
	err := dataBase.Cash.HSet(ctx, "product:"+productId, "likesOfProduct", likes).Err()
	if err != nil {
		return status.Errorf(codes.Internal, "unable to add data to hash of Cash  : %w", err)
	}
	return nil

}
