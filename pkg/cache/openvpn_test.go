package cache

import (
	"fmt"
	"testing"
	"time"
)

func Test_VPNClient_ListStatus(t *testing.T) {
	fmt.Println(time.Now().Unix())
	for v := range VPNClient.List("yunex") {
		if v == nil {
			break
		}
		fmt.Println(v)
	}
	for v := range VPNClient.List("guest") {
		if v == nil {
			break
		}
		fmt.Println(v)
	}
}
