# 클라이언트 모니터링 연동 가이드

## 개요
본 문서는 Fyne 등 클라이언트 애플리케이션에서 서버(라즈베리파이)의 상태를 모니터링하기 위한 API 연동 방법을 설명합니다.
현재 이 API는 HTTP Polling 방식에서 Server-Sent Events (SSE) 기반 스트리밍 방식으로 변경되었습니다.

## 엔드포인트
- **Method:** `GET`
- **Path:** `/api/v1/monitor/stream`
- **인증:** Tailscale 네트워크(VPN)를 통한 접근 필수
- **Content-Type:** `text/event-stream`

## 응답 구조 (SSE Event)
서버는 클라이언트가 연결을 유지하는 동안 주기적(약 2초 간격)으로 서버 측에서 측정된 `SystemStatus` JSON 데이터를 `message` 이벤트로 브로드캐스트합니다.

```http
event: message
data: {"cpuPercent":12.34,"memTotal":8345214976,"memUsed":4012314624,"memPercent":48.07,"temp":45.2}

event: message
data: {"cpuPercent":13.01,"memTotal":8345214976,"memUsed":4022314624,"memPercent":48.19,"temp":45.5}
```

### 필드 설명
| 필드명 | 타입 | 설명 |
|---|---|---|
| `cpuPercent` | `float64` | 전체 CPU 사용률 (%) |
| `memTotal` | `uint64` | 물리 메모리 전체 용량 (Bytes) |
| `memUsed` | `uint64` | 사용 중인 메모리 (Bytes) |
| `memPercent` | `float64` | 물리 메모리 사용률 (%) |
| `temp` | `float64` | 시스템 최고 온도 (섭씨, 섭씨 센서가 없으면 0.0) |

## 클라이언트(Fyne) 구현 시 참고사항
1. **SSE 연결 유지:** 지속적인 HTTP(또는 HTTPS) GET 요청을 보내고 응답의 본문(Body)을 스트리밍 방식으로 계속 읽어야 합니다. (예: `bufio.NewReader(resp.Body)`)
2. **이벤트 파싱:** `data: `로 시작하는 라인의 JSON 페이로드를 파싱하여 UI에 반영합니다.
3. **단위 변환:** `memTotal`, `memUsed`는 Bytes 단위이므로, 클라이언트에서 보기 좋게 GB/MB로 변환하여 표시해야 합니다.
4. **자동 재연결:** 네트워크 단절이나 에러 발생 시, 적절한 백오프(Backoff) 전략을 사용하여 엔드포인트에 다시 연결하도록 구현해야 합니다.
5. **에러 핸들링:** Tailscale 인증 실패 시 401/403 응답이 발생할 수 있습니다. 클라이언트 UI에 적절한 권한 에러 메시지를 노출하세요.
