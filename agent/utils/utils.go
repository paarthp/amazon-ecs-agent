package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/amazon-ecs-agent/agent/logger"
)

var log = logger.ForModule("util")

func DefaultIfBlank(str string, default_value string) string {
	if len(str) == 0 {
		return default_value
	}
	return str
}

func ZeroOrNil(obj interface{}) bool {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return true
	}
	if obj == nil {
		return true
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		return value.Len() == 0
	}
	zero := reflect.Zero(reflect.TypeOf(obj))
	if obj == zero.Interface() {
		return true
	}
	return false
}

// SlicesDeepEqual checks if slice1 and slice2 are equal, disregarding order.
func SlicesDeepEqual(slice1, slice2 interface{}) bool {
	s1 := reflect.ValueOf(slice1)
	s2 := reflect.ValueOf(slice2)

	if s1.Len() != s2.Len() {
		return false
	}
	if s1.Len() == 0 {
		return true
	}

	s2found := make([]int, s2.Len())
OuterLoop:
	for i := 0; i < s1.Len(); i++ {
		s1el := s1.Slice(i, i+1)
		for j := 0; j < s2.Len(); j++ {
			if s2found[j] == 1 {
				// We already counted this s2 element
				continue
			}
			s2el := s2.Slice(j, j+1)
			if reflect.DeepEqual(s1el.Interface(), s2el.Interface()) {
				s2found[j] = 1
				continue OuterLoop
			}
		}
		// Couldn't find something unused equal to s1
		return false
	}
	return true
}

func RandHex() string {
	randInt, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	out := make([]byte, 10)
	binary.PutVarint(out, randInt.Int64())
	return hex.EncodeToString(out)
}

func Strptr(s string) *string {
	return &s
}

// RetryWithBackoff takes a Backoff and a function to call that returns an error
// If the error is nil then the function will no longer be called
// If the error is Retriable then that will be used to determine if it should be
// retried
func RetryWithBackoff(backoff Backoff, fn func() error) {
	for err := fn(); true; err = fn() {
		retriable, isRetriable := err.(Retriable)

		if err == nil || isRetriable && !retriable.Retry() {
			return
		}

		time.Sleep(backoff.Duration())
	}
}

// Uint16SliceToStringSlice converts a slice of type uint16 to a slice of type
// *string. It uses strconv.Itoa on each element
func Uint16SliceToStringSlice(slice []uint16) []*string {
	stringSlice := make([]*string, len(slice))
	for i, el := range slice {
		str := strconv.Itoa(int(el))
		stringSlice[i] = &str
	}
	return stringSlice
}
