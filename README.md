# Network Storage (Tailscale 기반 중앙 통제 및 파일 공유 시스템)

이 프로젝트는 **Tailscale 가상 사설망(VPN)**을 기반으로 안전하게 구축된 중앙 집중형 파일 저장소 및 원격 제어 시스템입니다. Linux 서버를 백엔드로 사용하며, Fyne 프레임워크로 제작된 크로스 플랫폼 데스크톱 클라이언트를 통해 파일을 관리하고 시스템을 원격으로 모니터링/제어할 수 있습니다.

## 🚀 주요 기능

1. **강력한 보안 및 인증 (Tailscale Auth)**
   - 별도의 아이디/비밀번호 없이 Tailscale 망(Tailnet)에 접속된 기기와 사용자만 접근할 수 있습니다.
   - 백엔드는 Tailscale의 API를 사용하여 클라이언트의 신원(IP:Port)을 확실하게 검증하며 IP 스푸핑을 방지합니다.

2. **파일 관리 시스템 (HTTP API & SMB/NFS)**
   - 클라이언트 데스크톱 앱(UI)을 통해 손쉽게 파일 업로드, 다운로드, 폴더 생성, 삭제가 가능합니다. (대용량 파일 업로드 시 메모리 최적화를 위한 스트리밍 지원)
   - OS 자체의 탐색기를 선호하는 사용자를 위해 Tailscale 망 내에서만 접근 가능한 **Samba(SMB) 및 NFS** 마운트를 동시 지원합니다.

3. **원격 터미널 (WebTerminal)**
   - 데스크톱 클라이언트 앱 내에서 백엔드 서버의 쉘(Shell)에 직접 접근할 수 있는 PTY 기반 터미널 탭을 제공합니다. (WebSocket 사용)

4. **화면 공유 및 원격 제어 (WebRTC)**
   - 클라이언트 간 또는 클라이언트-서버 간의 실시간 화면 전송 및 스트리밍을 지원합니다.
   - 백엔드는 클라이언트들이 빠른 속도로 Peer-to-Peer(P2P) 직접 연결을 맺을 수 있도록 WebRTC 시그널링(Signaling) 서버 역할을 수행합니다.

5. **시스템 모니터링**
   - 중앙 서버의 CPU, 메모리, 디스크 사용량 등 하드웨어 상태를 클라이언트 대시보드에서 실시간으로 확인할 수 있습니다.

---

## 🛠️ 기술 스택 (Tech Stack)

### Backend (`backend/`)
- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Key Libraries**: `tailscale.com/client/tailscale`, `coder/websocket`, `creack/pty`
- **Features**: HTTP REST API, WebSocket Signaling/Terminal, System Monitor

### Client (`fyne-frontend/`)
- **Language**: Go (Golang)
- **Framework**: Fyne (Cross-platform GUI)
- **Key Libraries**: `pion/webrtc` (WebRTC 통신), `gorilla/websocket`
- **Features**: File Browser UI, Terminal UI, System Dashboard, WebRTC Viewer

---

## 📂 프로젝트 구조

```text
Network-Storage/
├── backend/                   # 중앙 통제 백엔드 서버 (라즈베리파이/Linux)
│   ├── cmd/server/            # 실행 진입점 (main.go)
│   ├── internal/              # 핵심 비즈니스 로직 (API, Middleware, WebRTC Signaling 등)
│   ├── scripts/               # 자동화 및 빌드 스크립트 (build.sh 등)
│   └── configs/               # 서버 환경 설정 파일 (app.json)
├── fyne-frontend/             # 데스크톱 클라이언트 애플리케이션 (GUI)
│   ├── internal/app/          # 클라이언트 앱 초기화 로직
│   ├── internal/client/       # 백엔드 통신용 HTTP/WebSocket 클라이언트 (업/다운로드 등)
│   ├── internal/ui/           # Fyne 프레임워크 기반 뷰(화면) 컴포넌트
│   └── internal/webrtc/       # 화면 전송 및 Peer 연결 로직
├── setting.md                 # ⚙️ 서버 및 클라이언트 환경 구축 매뉴얼
└── README.md                  # 프로젝트 소개서 (현재 문서)
```

---

## 📖 시작하기 (Getting Started)

이 서비스가 정상적으로 작동하기 위해서는 서버와 클라이언트 모두 **동일한 Tailscale 네트워크**에 연동되어 있어야 하며, 서버 환경에는 마운트 디렉터리 구성 및 시스템 데몬(SMB/NFS) 세팅이 필요합니다. 

구축 및 실행에 대한 가장 상세한 매뉴얼은 아래 문서를 반드시 먼저 참고하시기 바랍니다.

👉 **[환경 세팅 가이드 (setting.md) 바로가기](setting.md)**

### 요약: 백엔드 빌드 및 실행
```bash
cd backend
chmod +x scripts/build.sh
./scripts/build.sh
./bin/backend-server
```

### 요약: 클라이언트(프론트엔드) 실행
```bash
cd fyne-frontend
go mod tidy
go run main.go
# 앱 우측 상단의 톱니바퀴 ⚙️ 클릭 -> 서버의 Tailscale IP 입력 후 사용
```

---

## 🛡️ 네트워크 통신 아키텍처 (Architecture)

```mermaid
graph LR
    subgraph Tailscale VPN 
        Client[Desktop App\n(Fyne GUI)] <-->|HTTP/WS\nEncrypted| Backend[Backend Server\n(Gin API)]
        Client <-->|WebRTC P2P\nDirect Stream| Client2[Other Clients]
        Client <-->|SMB/NFS\nDirect Mount| BackendFS[Shared Directory]
    end
```
