package watcher

import (
	"fmt"
	"time"

	schedulepb "github.com/je0ng3/remindme-cli/api/proto/schedulepb"
	"github.com/je0ng3/remindme-cli/internal/notify"
)

type ScheduleChecker interface {
	Exists(id string) bool
	Delete(id string) error
}

func Watch(req *schedulepb.ScheduleRequest, checker ScheduleChecker) {
	layout := "2006-01-02 15:04"
	t, err := time.ParseInLocation(layout, req.Datetime, time.Local)
	if err != nil {
		fmt.Println("날짜 포맷 불일치:", err)
	}

	duration := time.Until(t)
	if duration <= 0 {
		return
	}
	
	time.Sleep(duration)

	if checker.Exists(req.Id) {
		err = notify.Send(req.Title, req.Memo, req.Url)
		if err != nil {
			fmt.Println("알림 전송 실패:", err)
		}
		checker.Delete(req.Id)
	}
}