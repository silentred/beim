package lib

import (
	"fmt"
)

func RedisSessKey(username string) string {
	return fmt.Sprintf("sess:%s", username)
}

func RedisOfflineMsgKey(username string) string {
	return fmt.Sprintf("ofm:%s", username)
}

func RedisCometTopic(cometID string) string {
	return fmt.Sprintf("cmt/%s", cometID)
}
