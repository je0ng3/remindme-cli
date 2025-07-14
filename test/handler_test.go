package test

import (
	"context"
	"os"
	"testing"

	"github.com/je0ng3/remindme-cli/api/proto/schedulepb"
	"github.com/je0ng3/remindme-cli/internal/server"
)

func createTempServer(t *testing.T) (*server.ScheduleServer, string, func()) {
	tmpFile, err := os.CreateTemp("", "schedule_test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	s := server.NewSchedulerServer(tmpFile.Name())
	return s, tmpFile.Name(), cleanup
}

func TestAddSchedule_Success(t *testing.T) {
	s, _, cleanup := createTempServer(t)
	defer cleanup()

	ctx := context.TODO()

	_, err := s.AddSchedule(ctx, &schedulepb.ScheduleRequest{
		Title:    "Test Schedule",
		Datetime: "2025-07-20 10:00",
		Url:      "https://example.com",
		Memo:     "Test memo",
	})
	if err != nil {
		t.Fatalf("AddSchedule failed: %v", err)
	}
}

func TestAddSchedule_TitleRequired(t *testing.T) {
	s, _, cleanup := createTempServer(t)
	defer cleanup()

	ctx := context.TODO()

	_, err := s.AddSchedule(ctx, &schedulepb.ScheduleRequest{
		Title: "", // intentionally left empty
	})
	if err == nil {
		t.Fatal("expected error when title is missing, got nil")
	}
}

func TestListSchedules(t *testing.T) {
	s, _, cleanup := createTempServer(t)
	defer cleanup()

	ctx := context.TODO()

	// Add multiple schedules
	s.AddSchedule(ctx, &schedulepb.ScheduleRequest{
		Title:    "Schedule One",
		Datetime: "2025-07-20",
	})
	s.AddSchedule(ctx, &schedulepb.ScheduleRequest{
		Title:    "Schedule Two",
		Datetime: "2025-07-21",
	})

	resp, err := s.ListSchedules(ctx, &schedulepb.Empty{})
	if err != nil {
		t.Fatalf("ListSchedules failed: %v", err)
	}

	if len(resp.Schedules) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(resp.Schedules))
	}
}

func TestDeleteSchedule_Success(t *testing.T) {
	s, _, cleanup := createTempServer(t)
	defer cleanup()

	ctx := context.TODO()

	// Add one schedule
	_, err := s.AddSchedule(ctx, &schedulepb.ScheduleRequest{
		Title:    "Delete Me",
		Datetime: "2025-07-22",
	})
	if err != nil {
		t.Fatalf("AddSchedule failed: %v", err)
	}

	// Delete it
	delResp, err := s.DeleteSchedule(ctx, &schedulepb.ScheduleIdx{Idx: 1})
	if err != nil {
		t.Fatalf("DeleteSchedule failed: %v", err)
	}
	if delResp.Message != "Schedule deleted." {
		t.Errorf("Unexpected delete message: %s", delResp.Message)
	}

	// Verify it's gone
	listAfter, _ := s.ListSchedules(ctx, &schedulepb.Empty{})
	if len(listAfter.Schedules) != 0 {
		t.Errorf("Expected 0 schedules after deletion, got %d", len(listAfter.Schedules))
	}
}

func TestDeleteSchedule_NotFound(t *testing.T) {
	s, _, cleanup := createTempServer(t)
	defer cleanup()

	ctx := context.TODO()

	resp, err := s.DeleteSchedule(ctx, &schedulepb.ScheduleIdx{Idx: 999})
	if err != nil {
		t.Fatalf("DeleteSchedule returned error: %v", err)
	}
	if resp.Message != "Invalid index" {
		t.Errorf("Expected 'Invalid index', got '%s'", resp.Message)
	}
}