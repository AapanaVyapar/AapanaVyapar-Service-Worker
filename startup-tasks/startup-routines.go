package services

import (
	"context"
	"fmt"
	"strings"
)

type DataSources struct {
	Data *MongoDataBase
	Cash *RedisDataBase
}

func NewDataSource() *DataSources {
	return &DataSources{
		Data: NewMongoClient(),
		Cash: NewRedisClient(),
	}
}

func (dataSource *DataSources) InitCartStream(ctx context.Context) error {
	err := dataSource.Cash.CreateCartStream(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			fmt.Println("Cart Already Exist")
			return nil
		}
		return err
	}
	return nil
}

func (dataSource *DataSources) InitFavStream(ctx context.Context) error {
	err := dataSource.Cash.CreateFavStream(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			fmt.Println("Fav Already Exist")
			return nil
		}
		return err
	}
	return nil
}
