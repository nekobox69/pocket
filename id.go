// Package pocket Create at 2020-11-06 10:18
package pocket

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sony/sonyflake"
)

var (
	mutex sync.Mutex
	st    sonyflake.Settings
	sf    *sonyflake.Sonyflake
)

// GetUUID gene uuid
func GetUUID() uuid.UUID {
	mutex.Lock()
	defer mutex.Unlock()
	return uuid.NewV4()
}

// GetUUIDStr gene uuid
func GetUUIDStr() string {
	return fmt.Sprintf("%s", GetUUID())
}

// CreateNumberCaptcha random
func CreateNumberCaptcha(digit int) string {
	format := "%0" + fmt.Sprintf("%d", digit) + "v"
	return fmt.Sprintf(format, rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(Pow(10, int64(digit))))
}

// SonyFlakeId SonyFlake id
func SonyFlakeId() (uint64, error) {
	return sf.NextID()
}
