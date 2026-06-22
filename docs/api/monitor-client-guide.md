# 클라이언트 모니터링 연동 가이드

## 개요
본 문서는 Fyne 등 클라이언트 애플리케이션에서 서버(라즈베리파이)의 상태를 모니터링하기 위한 API 연동 방법을 설명합니다.

## 엔드포인트
- **Method:** `GET`
- **Path:** `/api/v1/monitor`
- **인증:** Tailscale 네트워크(VPN)를 통한 접근 필수

## 응답 구조 (JSON)
서버는 다음과 같은 `SystemStatus` JSON 구조를 반환합니다. 클라이언트 측에서 파싱할 때 참고하세요.

```json
{
  "cpuPercent": 12.34,
  "memTotal": 8345214976,
  "memUsed": 4012314624,
  "memPercent": 48.07,
  "temp": 45.2
}
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
1. **타이머/주기적 폴링:** Fyne 대시보드에서는 `time.Ticker`를 사용하여 1초~5초 간격으로 위 API를 호출하여 화면을 갱신하는 것을 권장합니다.
2. **단위 변환:** `memTotal`, `memUsed`는 Bytes 단위이므로, 클라이언트에서 보기 좋게 GB/MB로 변환하여 표시해야 합니다.
3. **에러 핸들링:** Tailscale 인증 실패 시 401/403 응답이 발생할 수 있습니다. 클라이언트 UI에 적절한 권한 에러 메시지를 노출하세요.
