package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.foxminded.ua/foxstudent106264/task-3.5/cmd/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client         *redis.Client
	expTimeSeconds time.Duration
}

func NewRedis(addr string, db int, expTime int) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       db,
	})
	return &Redis{Client: client, expTimeSeconds: time.Duration(expTime) * time.Second}
}

func (r *Redis) Set(key string, value interface{}) error {
	json, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("Set: unable to marshall JSON: %w", err)
	}
	r.Client.Set(context.Background(), key, json, r.expTimeSeconds)
	return nil
}

func (r *Redis) GetUser(key string) (domain.UserProfileDTO, error) {
	res, err := r.Client.Get(context.Background(), key).Result()
	if err != nil || res == "" {
		return domain.UserProfileDTO{}, err
	}
	var user domain.UserProfileDTO
	err = json.Unmarshal([]byte(res), &user)
	if err != nil {
		return domain.UserProfileDTO{}, fmt.Errorf("getUser: unable to decode JSON: %w", err)
	}
	return user, nil
}

func (r *Redis) GetUsersList(key string) (domain.Pagination[domain.UserProfileDTO], error) {
	res, err := r.Client.Get(context.Background(), key).Result()
	if err != nil || res == "" {
		return domain.Pagination[domain.UserProfileDTO]{}, err
	}

	var usersList domain.Pagination[domain.UserProfileDTO]
	err = json.Unmarshal([]byte(res), &usersList)
	if err != nil {
		return domain.Pagination[domain.UserProfileDTO]{}, fmt.Errorf("getUser: unable to decode JSON: %w", err)
	}
	return usersList, nil
}
func (r *Redis) MakeKey(pageSize int, offset int) string {
	return fmt.Sprintf("pageSize:%d,offset:%d", pageSize, offset)
}
