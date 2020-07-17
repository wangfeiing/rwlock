package tool

import (
	"github.com/sony/sonyflake"
	"strconv"
)

var sonyflakeGen = sonyflake.NewSonyflake(sonyflake.Settings{})

func GetUUID() string {
	uuid, _ := sonyflakeGen.NextID()
	return strconv.Itoa(int(uuid))
}
