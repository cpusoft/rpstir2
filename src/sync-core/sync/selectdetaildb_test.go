package sync

import (
	"fmt"
	"testing"
	"time"
)

func TestInsertSelectDetailDb(t *testing.T) {
	logID := "logid"
	index := int(0)
	startTime := time.Now()
	detail := "{}"
	gotSyncLogId, err := InsertSelectDetailDb(logID, index, startTime, detail)
	fmt.Println(gotSyncLogId)
	fmt.Println(err)

}
