package server

import (
	"encoding/csv"
	"os"
	"sync"

	"context"

	schedulepb "github.com/je0ng3/remindme-cli/api/proto/schedulepb"
)


type ScheduleServer struct {
	schedulepb.UnimplementedSchedulerServer
	mu			sync.Mutex
	csvFile		string
}


func NewSchedulerServer(csvPath string) *ScheduleServer {
	return &ScheduleServer{
		csvFile: csvPath,
	}
}

func (s *ScheduleServer) AddSchedule(ctx context.Context, req *schedulepb.ScheduleRequest) (*schedulepb.ScheduleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{req.Id, req.Title, req.Datetime, req.Url, req.Memo})
	if err != nil {
		return nil, err
	}

	return &schedulepb.ScheduleResponse{Message: "Schedule added."}, nil
}



