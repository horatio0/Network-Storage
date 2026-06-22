# Tailscale API 가이드

이 문서는 백엔드에서 제공하는 Tailscale 관련 API의 사용법을 설명합니다.

## 주의 사항 (사전 요구 조건)

이 API는 서버가 실행되는 환경의 `tailscaled` 로컬 소켓(예: `/var/run/tailscale/tailscaled.sock`)과 직접 통신하여 상태 정보를 조회합니다. 
따라서 백엔드 애플리케이션을 실행하는 사용자 계정이 **root 권한**을 가지거나, **tailscale 그룹**에 속해 있어서 소켓 파일에 접근할 수 있어야 합니다.

소켓 접근 권한이 없을 경우 API 호출 시 `500 Internal Server Error` (failed to get tailscale status)가 반환될 수 있습니다.

---

## 1. Tailscale 연결 장치 목록 조회

Tailscale 네트워크(Tailnet)에 연결된 전체 Peer(장치) 목록과 각 장치의 운영체제(OS) 정보를 반환합니다. 
이를 통해 클라이언트는 해당 장치가 리눅스인지 윈도우인지 판단하고, 적절한 접속 방식(터미널 또는 원격 화면)을 선택할 수 있습니다.

**Endpoint:** `GET /api/v1/tailscale/devices`

### 요청 (Request)

추가적인 Query Parameter나 Body는 필요하지 않습니다.

```http
GET /api/v1/tailscale/devices HTTP/1.1
Host: your-backend-domain.com
```

### 응답 (Response)

- **200 OK**: 성공적으로 장치 목록을 가져온 경우

```json
{
  "devices": [
    {
      "name": "desktop-windows.local.domain",
      "ips": [
        "100.101.102.103",
        "fd7a:115c:a1e0:ab12:4843:cd96:6265:6667"
      ],
      "os": "windows"
    },
    {
      "name": "server-linux.local.domain",
      "ips": [
        "100.111.112.113",
        "fd7a:115c:a1e0:ab12:4843:cd96:6265:6668"
      ],
      "os": "linux"
    }
  ]
}
```

- **500 Internal Server Error**: Tailscale 로컬 소켓에 접근할 수 없거나 내부 상태 조회에 실패한 경우

```json
{
  "error": "failed to get tailscale status"
}
```

### 필드 설명

- `name` (string): 장치의 호스트명(HostName) 또는 DNS명입니다. (DNS명이 있으면 우선 사용됩니다.)
- `ips` (array of string): 해당 장치에 할당된 Tailscale IP 주소 목록입니다. (IPv4, IPv6 포함)
- `os` (string): 장치의 운영체제 정보입니다. 주로 `linux`, `windows`, `macOS`, `iOS`, `android` 등의 문자열이 반환됩니다.
