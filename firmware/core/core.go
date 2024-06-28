package core

import (
	"fmt"
	"time"
)

func Failed(err error) {
	for {
		fmt.Println("[ERR]", err.Error())
		time.Sleep(3 * time.Second)
	}
}
