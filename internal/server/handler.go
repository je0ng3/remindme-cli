package server

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/google/uuid"
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

	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	id := uuid.New().String()

	file, err := os.OpenFile(s.csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{id, req.Title, req.Datetime, req.Url, req.Memo})
	if err != nil {
		return nil, err
	}

	return &schedulepb.ScheduleResponse{Message: "Schedule added."}, nil
}


func (s *ScheduleServer) ListSchedules(ctx context.Context, _ *schedulepb.Empty) (*schedulepb.ScheduleList, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.csvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var list []*schedulepb.ScheduleRequest
	for _, r := range records {
		list = append(list, &schedulepb.ScheduleRequest{
			Id:			r[0],
			Title: 		r[1],
			Datetime: 	r[2],
			Url: 		r[3],
			Memo: 		r[4],
		})
	}

	fmt.Println("Schedules:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tDatetime\tURL\tMemo")
	for _, sch := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", sch.Id, sch.Title, sch.Datetime, sch.Url, sch.Memo)
	}
	w.Flush()

	return &schedulepb.ScheduleList{Schedules: list}, nil
}

func (s *ScheduleServer) DeleteSchedule(ctx context.Context, req *schedulepb.ScheduleId) (*schedulepb.ScheduleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.csvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var updated [][]string
	deleted := false
	for _, r := range records {
		if r[0] == req.Id {
			deleted = true
			continue
		}
		updated = append(updated, r)
	}
	if !deleted {
		return &schedulepb.ScheduleResponse{Message: "Schedule not found."}, nil
	}

	file, err = os.Create(s.csvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.WriteAll(updated)

	return &schedulepb.ScheduleResponse{Message: "Schedule deleted."}, nil
}
