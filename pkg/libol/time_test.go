package libol

import (
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	ti, _ := time.Parse(time.UnixDate, "Mon Nov 30 21:45:49 CST 2020")
	fmt.Println(ti)
	fmt.Println(time.Since(ti))
	fmt.Println(time.Now())
}
