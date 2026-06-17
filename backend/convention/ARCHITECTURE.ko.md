# 아키텍처 가이드

## 핵심 설계 원칙

이 프로젝트는 **보안**과 **기능별 독립성**을 최우선으로 한다.

### 1. Tailscale 기반 제로 트러스트
모든 API 요청은 `internal/middleware/tailscale.go`를 거치며, Tailscale 네트워크 내의 인가된 사용자인지 확인한다.

### 2. 설정 중심 부팅
`configs/app.json`에 정의된 설정을 바탕으로 서버가 구성된다. `internal/config` 패키지가 이를 로드하고 검증한다.

### 3. 기능별 격리
모니터링, 파일 전송 등 각 기능은 `internal/` 하위의 독립된 패키지로 구현되어 서로 간섭하지 않는다.

## 요청 처리 흐름

1. **Request:** 클라이언트가 API 호출
2. **Middleware:** 
   - `Logger`: 요청 로깅
   - `Recovery`: 패닉 복구
   - `TailscaleAuth`: 사용자 인증 및 차단
3. **Routing:** `internal/app/app.go`에서 정의된 경로에 따라 핸들러 매칭
4. **Execution:** 각 패키지(`monitor`, `files` 등)의 핸들러가 로직 수행
5. **Response:** JSON 결과 반환

## 향후 확장 계획
- **WoL:** 내부망 브로드캐스트를 통한 전원 제어 추가
- **Signaling:** WebRTC 연결을 위한 시그널링 엔드포인트 추가
