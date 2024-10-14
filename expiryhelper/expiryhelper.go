package expiryhelper

import (
	"log"
	"sync"
	"time"
)

// Declare the timeMap and a mutex for thread safety
var pinExpiryMap = make(map[string]time.Time)
var mu sync.Mutex

func GetPinExpiryMap() map[string]time.Time {
	return pinExpiryMap
}

func AddPinExpiry(pin string, hours int) {
	mu.Lock()
	defer mu.Unlock()
	futureTime := time.Now().Add(time.Duration(hours) * time.Hour)

	log.Printf("Adding pin %s with expiry time %s", pin, futureTime)
	pinExpiryMap[pin] = futureTime
}

func SetPinExpiry(pin string, expiryTime time.Time) {
	mu.Lock()
	defer mu.Unlock()
	pinExpiryMap[pin] = expiryTime
}

func GetPinExpiry(pin string) (time.Time, bool) {
	mu.Lock()
	defer mu.Unlock()
	value, exists := pinExpiryMap[pin]
	return value, exists
}

// RemoveTime deletes a pin from the pinExpiryMap
func RemovePinExpiry(pin string) {
	mu.Lock()
	defer mu.Unlock()
	delete(pinExpiryMap, pin)
}

func IsPinExpired(pin string) bool {
	expiryTime, exists := GetPinExpiry(pin)
	if !exists {
		return true
	}
	return time.Now().After(expiryTime)
}
