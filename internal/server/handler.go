package server

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"
	"time"

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
	req.Id = id

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
	go s.WatchSchedule(req)
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	w.Flush()

	return &schedulepb.ScheduleList{Schedules: list}, nil
}

func (s *ScheduleServer) DeleteSchedule(ctx context.Context, req *schedulepb.ScheduleIdx) (*schedulepb.ScheduleResponse, error) {
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

	idx := int(req.Idx) - 1
	if idx < 0 || idx >= len(records) {
		return &schedulepb.ScheduleResponse{Message: "Invalid index"}, nil
	}
	records = append(records[:idx], records[idx+1:]...)

	updatedFile, err := os.Create(s.csvFile)
	if err != nil {
		return nil, err
	}
	defer updatedFile.Close()

	writer := csv.NewWriter(updatedFile)
	defer writer.Flush()
	
	err = writer.WriteAll(records)
	if err != nil {
		return nil, err
	}

	return &schedulepb.ScheduleResponse{Message: "Schedule deleted."}, nil
}


func (s *ScheduleServer) WatchSchedule (req *schedulepb.ScheduleRequest) {
	layout := "2006-01-02 15:04"
	loc := time.Local
	t, err := time.ParseInLocation(layout, req.Datetime, loc)
	if err != nil {
		fmt.Println("날짜 포맷 불일치:", err)
	}

	duration := time.Until(t)
	if duration <= 0 {
		return
	}
	
	time.Sleep(duration)

	if s.scheduleExists(req.Id) {
		err = notification(req.Title, req.Memo, req.Url)
		if err != nil {
			fmt.Println("알림 전송 실패:", err)
		}
		err = s.deleteExpiredSchedule(req.Id)
		if err != nil {
			fmt.Println("일정 삭제 실패:", err)
		}
	}
}

func notification(title, memo, url string) error {
	args := []string {"-title", title}
	if memo != "" {
		args = append(args, "-message", memo)
	}
	if url != "" {
		args = append(args, "-open", url)
	}
	fmt.Println(args)
	cmd := exec.Command("terminal-notifier", args...)
	return cmd.Run()
}


func (s *ScheduleServer) deleteExpiredSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	var updatedRecords [][]string
	for _, record := range records {
		if record[0] != id {
			updatedRecords = append(updatedRecords, record)
		}
	}

	updatedFile, err := os.Create(s.csvFile)
	if err != nil {
		return err
	}
	defer updatedFile.Close()

	writer := csv.NewWriter(updatedFile)
	defer writer.Flush()
	return writer.WriteAll(updatedRecords)
}

func (s *ScheduleServer) scheduleExists(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.csvFile)
	if err != nil {
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return false
	}

	for _, record := range records {
		if record[0] == id {
			return true
		}
	}
	return false
}