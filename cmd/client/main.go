package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	schedulepb "github.com/je0ng3/remindme-cli/api/proto/schedulepb"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: remindme add | list | delete [index]")
		return 
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := schedulepb.NewSchedulerClient(conn)

	switch os.Args[1] {
	case "add":
		runAddCommand(client)
	case "list":
		runListCommand(client)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("삭제할 인덱스를 입력하세요.")
		}
		runDeleteCommand(client, os.Args[2])
	default:
		fmt.Println("사용법: remindme add | list | delete [index]")
	}
}

func runAddCommand(client schedulepb.SchedulerClient) {
	tmpfile, err := os.CreateTemp("", "remindme_*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	template := `# 템플릿에 맞춰 일정 정보를 입력하세요. Title 및 Datetime은 필수입니다.
	Title:
	Datetime: 2003-03-01 07:30
	URL:
	Memo:
	`
	
	if _, err := tmpfile.Write([]byte(template)); err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("에디터 실행 실패: %v", err)
	}

	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	title, datetime, url, memo := "", "", "", ""
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		} else if strings.HasPrefix(line, "Datetime:") {
			datetime =strings.TrimSpace(strings.TrimPrefix(line, "Datetime:"))
		} else if strings.HasPrefix(line, "URL:") {
			url =strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
		} else if strings.HasPrefix(line, "Memo:") {
			memo =strings.TrimSpace(strings.TrimPrefix(line, "Memo:"))
		}
	}

	if title == "" || datetime == "" {
		fmt.Println("Title과 Datetime은 필수입니다.")
		return
	}

	req := &schedulepb.ScheduleRequest{
		Title:    title,
		Datetime: datetime,
		Url:      url,
		Memo:     memo,
	}

	res, err := client.AddSchedule(context.Background(), req)
	if err != nil {
		fmt.Println("등록 실패:", err)
	} else {
		fmt.Println("일정 추가됨:", res.Message)
	}
}

func runListCommand(client schedulepb.SchedulerClient) {
	res, err := client.ListSchedules(context.Background(), &schedulepb.Empty{})
	if err != nil {
		fmt.Println("일정 목록 불러오기 실패:", err)
		return
	}

	if len(res.Schedules) == 0 {
		fmt.Println("등록된 일정이 없습니다.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "No\tTitle\tDatetime\tURL\tMemo")
	for i, sch := range res.Schedules {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i+1, sch.Title, sch.Datetime, sch.Url, sch.Memo)
	}
	w.Flush()
}

func runDeleteCommand(client schedulepb.SchedulerClient, idxArg string) {
	idx, err := strconv.Atoi(idxArg)
	if err != nil || idx <= 0 {
		fmt.Println("유효한 숫자 인덱스를 입력하세요")
		return
	}

	req := &schedulepb.ScheduleIdx{Idx: int32(idx)}
	res, err := client.DeleteSchedule(context.Background(), req)
	if err != nil {
		fmt.Println("삭제 요청 실패:", err)
		return
	}

	if res.Message == "Invalid index" {
		fmt.Println("존재하지 않는 인덱스입니다.")
	} else {
		fmt.Println("일정삭제 완료", res.Message)
	}
}