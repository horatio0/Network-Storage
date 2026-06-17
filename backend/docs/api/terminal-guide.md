# 터미널 API 가이드

이 문서는 `central-control-backend`에서 제공하는 원격 터미널 WebSocket API의 연동 방법을 설명한다.

## 엔드포인트
- **URL:** `ws://<server-ip>:<port>/api/v1/terminal/ws`
- **프로토콜:** WebSocket (Binary Message)
- **보안:** Tailscale 네트워크망 접근 및 인증 필요

## 통신 방식

### 1. 연결 (Connection)
클라이언트가 `/api/v1/terminal/ws`로 WebSocket 업그레이드 요청을 보내면, 서버는 내부적으로 `/bin/bash`(또는 `/bin/sh`) 프로세스를 PTY와 함께 실행한다.

### 2. 데이터 송신 (Input)
클라이언트에서 입력한 키보드 이벤트나 명령어 문자열을 **Binary 메시지** 형태로 보낸다. 서버는 이를 PTY의 표준 입력(stdin)으로 전달한다.

### 3. 데이터 수신 (Output)
서버는 PTY의 표준 출력(stdout) 및 표준 에러(stderr)에서 발생하는 데이터를 실시간으로 **Binary 메시지** 형태로 클라이언트에 보낸다. 클라이언트는 이를 터미널 에뮬레이터 UI에 렌더링해야 한다.

### 4. 연결 종료 (Disconnection)
WebSocket 연결이 끊어지면 서버는 실행 중이던 셸 프로세스를 즉시 종료하고 자원을 회수한다.

## 클라이언트 구현 참고사항 (Fyne)
- `fyne.io/x/fyne-term` 등 기존 터미널 위젯 라이브러리를 활용하는 것이 권장된다.
- **TERM 환경변수:** 서버는 기본적으로 `xterm-256color` 환경변수를 사용하여 셸을 실행한다. 클라이언트의 터미널 위젯이 이를 지원해야 색상 등이 정상적으로 표시된다.
- **윈도우 크기 조절 (Resize):** 현재 버전에서는 고정된 터미널 크기를 사용한다. 향후 JSON 메시지 규격을 도입하여 크기 조절 기능을 추가할 예정이다.
