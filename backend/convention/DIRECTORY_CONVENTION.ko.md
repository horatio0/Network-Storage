# 디렉토리 컨벤션

## 목적

이 문서는 중앙 통제 컴퓨터 백엔드(`central-control-backend`)의 디렉토리 구조와 패키지 책임을 설명한다.

## 프로젝트 목표

이 저장소는 여러 장치를 중앙에서 제어하고 모니터링하기 위한 서버다.

주요 기능:
- **Tailscale VPN 기반 보안:** 인가된 사용자만 접근 허용
- **시스템 모니터링:** 서버 리소스(CPU, 메모리, 온도) 상태 제공
- **파일 제어:** 원격 파일 업로드 및 다운로드
- **WoL(Wake-on-LAN):** 데스크탑 전원 제어 (예정)
- **원격 제어 중계:** WebRTC 시그널링 및 터미널 접속 (예정)

## 최상위 구조

### `main.go`
실행 진입점. 설정을 로드하고 `internal/app`을 통해 서버를 실행한다.

### `configs/`
애플리케이션 설정 파일 보관.
- `app.json`: 서버 포트, 경로, Tailscale 설정 등.

### `convention/`
아키텍처 및 코딩 컨벤션 문서 보관.

### `docs/`
API 가이드 및 프로젝트 계획서 보관.

### `internal/`
실제 비즈니스 로직 및 기능 구현체. 외부에서 직접 임포트할 수 없도록 격리됨.

## 패키지별 책임

### `internal/app`
애플리케이션 조립 및 서버(Gin) 실행 관리. 라우팅 설정이 여기서 이루어짐.

### `internal/config`
설정 파일(`app.json`) 로드 및 검증.

### `internal/middleware`
공통 HTTP 미들웨어 (로깅, Tailscale 인증 등).

### `internal/monitor`
서버 리소스 상태 수집 로직.

### `internal/files`
파일 업로드 및 다운로드 처리 로직.

## 의존성 규칙
- 기능별 패키지(`monitor`, `files` 등)는 서로 독립적이어야 한다.
- 모든 기능의 조립은 `internal/app`에서 담당한다.
- `internal/config`는 설정을 제공할 뿐, 다른 기능 로직에 의존하지 않는다.
