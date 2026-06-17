# 타입 레퍼런스

## 목적
이 문서는 `central-control-backend` 프로젝트에 정의된 주요 타입들의 역할과 구조를 설명한다.

## 타입 분류

### 1. 설정 타입 (`internal/config`)
애플리케이션 구동에 필요한 설정을 담는다.

#### `AppConfig`
- **역할:** `configs/app.json` 파일의 Go 표현체.
- **주요 필드:**
  - `ListenAddr`: 서버가 리스닝할 주소 (예: `:8080`)
  - `MountPath`: 파일 전송 시 사용할 베이스 디렉토리
  - `Tailscale`: Tailscale 활성화 여부 및 허용 사용자 목록

### 2. 모니터링 타입 (`internal/monitor`)
시스템 리소스 상태를 표현한다.

#### `SystemStatus`
- **역할:** CPU, 메모리, 온도 정보를 담아 클라이언트에 전달하는 구조체.
- **주요 필드:**
  - `cpuPercent`: CPU 사용률
  - `memTotal`, `memUsed`, `memPercent`: 메모리 정보
  - `temp`: 시스템 온도 (Celsius)

### 3. 파일 제어 타입 (`internal/files`)
파일 업로드 및 다운로드 관련 정보를 다룬다. 주로 Gin의 `Context`를 통해 처리되나, 향후 파일 목록 조회 등을 위한 타입이 추가될 수 있다.

## 향후 추가 예정 타입

### WoL (Wake-on-LAN)
- `WoLTarget`: 전원을 켤 대상 기기의 이름, MAC 주소, 브로드캐스트 IP 등을 담는 타입.

### Signaling (WebRTC)
- `SignalMessage`: 클라이언트 간 WebRTC 연결을 위한 Offer, Answer, ICE Candidate 정보를 담는 타입.
