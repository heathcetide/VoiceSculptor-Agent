package util

import (
	"VoiceSculptor/pkg/logger"
	"encoding/base64"
	"errors"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var SnowflakeUtil *Snowflake
var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
var numberRunes = []rune("0123456789")

func init() {
	rand.Seed(time.Now().UnixNano())
	SnowflakeUtil, _ = NewSnowflake()
}

func randRunes(n int, source []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}

func RandText(n int) string {
	return randRunes(n, letterRunes)
}

func RandNumberText(n int) string {
	return randRunes(n, numberRunes)
}

func SafeCall(f func() error, failHandle func(error)) error {
	defer func() {
		if err := recover(); err != nil {
			if failHandle != nil {
				eo, ok := err.(error)
				if !ok {
					es, ok := err.(string)
					if ok {
						eo = errors.New(es)
					} else {
						eo = errors.New("unknown error type")
					}
				}
				failHandle(eo)
			} else {
				logger.Error("panic", zap.Any("error", err))
			}
		}
	}()
	return f()
}

func StructAsMap(form any, fields []string) (vals map[string]any) {
	vals = make(map[string]any)
	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return vals
	}
	for i := 0; i < len(fields); i++ {
		k := v.FieldByName(fields[i])
		if !k.IsValid() || k.IsZero() {
			continue
		}
		if k.Kind() == reflect.Ptr {
			if !k.IsNil() {
				vals[fields[i]] = k.Elem().Interface()
			}
		} else {
			vals[fields[i]] = k.Interface()
		}
	}
	return vals
}

// GenerateSecureToken 生成固定长度的安全 token
func GenerateSecureToken(length int) (string, error) {
	token := make([]byte, length)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

const (
	epoch         int64 = 1609459200000000 // 微秒级起始时间戳（2021-01-01）
	timestampBits uint  = 44
	machineIDBits uint  = 10
	sequenceBits  uint  = 9

	maxMachineID = -1 ^ (-1 << machineIDBits) // 1023
	maxSequence  = -1 ^ (-1 << sequenceBits)  // 511

	machineIDShift = sequenceBits
	timestampShift = machineIDBits + sequenceBits
)

type Snowflake struct {
	mu        sync.Mutex
	lastStamp int64
	sequence  int64
	machineID int64
}

func NewSnowflake() (*Snowflake, error) {
	id := getMachineID()
	if id < 0 || id > maxMachineID {
		return nil, errors.New("machineID 超出范围")
	}
	return &Snowflake{
		machineID: id,
	}, nil
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := currentMicro()
	if now < s.lastStamp {
		// 时钟回拨保护
		return 0
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// 当前微秒内序号已满，等待下一个微秒
			for now <= s.lastStamp {
				now = currentMicro()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = now

	id := ((now - epoch) << timestampShift) |
		(s.machineID << machineIDShift) |
		s.sequence

	return id
}

func currentMicro() int64 {
	return time.Now().UnixNano() / 1e3
}

func getMachineID() int64 {
	val := os.Getenv("MACHINE_ID")
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 1 // fallback 默认值，建议根据实际修改
	}
	return id
}
