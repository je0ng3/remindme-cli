# ⏰ remindme-cli 
일정 정보를 추가하고, 특정 시간에 알림을 받을 수 있는 gRPC 기반 CLI툴  
간단한 명령어로 일정을 관리하고 알림과 함께 메모를 전달받을 수 있으며, url도 전달할 시 알람 클릭을 통해 해당 페이지를 열 수 있음

### 기능
add: 일정추가
list: 일정 목록 조회
delete [index]: 일정 삭제
알람 시간 도래 시 macOS 알림 전송 (terminal-notifier 사용)
url 자동 열기 기능 포함

### 설치
```
git clone https://github.com/je0ng3/remindme-cli.git
cd remindme-cli
go build -o remindme cmd/client/main.go
go build -o remindserver cmd/server/main.go
```

### 사용법
서버 실행
```
./remindserver
```
일정 추가 - nano 편집기가 켜지면 아래 템플릿에 맞춰 작성
```
title: 회의
datetime: 2025-07-22 18:00
memo: 프로젝트 리뷰 회의
url: https://zoom.us/meeting/123
```
일정 목록
```
./remindcli list
```
일정 삭제 - 일정 목록에 있는 인덱스에 맞춰 작성
```
./remindcli delete
```

### + 전역 명령어로 사용
개인 bin 디렉토리로 이동시키기
```
mkdir -p ~/bin
mv remindcli ~/bin/
```

zsh에서 ~/bin이 $PATH에 없으면 추가
```
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

이제 전역 명령어로 실행 가능
```
remindcli list
```
