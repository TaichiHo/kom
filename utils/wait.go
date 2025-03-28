package utils

import (
	"time"

	"k8s.io/klog/v2"
)

// WaitUntil takes a function f and waits until f returns true or timeout is reached
// Example:
// // Define a check function that simulates a condition being met every 10 seconds
//
//	checkCondition := func() bool {
//		// For example, condition is true when current second is a multiple of 10
//		return time.Now().Second()%10 == 0
//	}
//
//	// Check every 2 seconds with a timeout of 30 seconds
//	interval := 2 * time.Second
//	timeout := 30 * time.Second
//
//	if waitUntil(checkCondition, interval, timeout) {
//		fmt.Println("Check succeeded, main process exiting.")
//	} else {
//		fmt.Println("Check failed due to timeout.")
//	}
func WaitUntil(f func() bool, interval time.Duration, timeout time.Duration) bool {
	timeoutTimer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer:
			klog.V(4).Infof("Timeout reached, stopping monitoring.")
			return false // Return false on timeout
		case <-ticker.C:
			if f() {
				klog.V(4).Infof("Condition met, stopping monitoring.")
				return true // Stop if f returns true
			}
			klog.V(2).Infof("Condition not met, retrying...")
		}
	}
}
