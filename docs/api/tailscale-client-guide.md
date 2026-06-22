# 클라이언트 보안(Tailscale) 연동 가이드

## 개요
중앙 통제 컴퓨터(서버)의 모든 API는 내부적으로 Tailscale Whois API를 사용하여 요청자의 신원을 검증합니다. 따라서 Fyne으로 제작될 클라이언트 애플리케이션은 반드시 서버와 동일한 Tailscale 네트워크(Tailnet) 상에서 통신해야 합니다.

이 문서는 클라이언트 측에서 서버 API를 호출하기 위해 알아야 할 아키텍처 및 구현 방식을 안내합니다.

## 통신 제약 사항
1. **IP 주소:** 클라이언트는 서버의 공인 IP나 로컬 망 IP가 아닌, 서버의 **Tailscale IP (100.x.x.x)** 또는 **MagicDNS 도메인**을 엔드포인트로 사용해야 합니다.
2. **인증 헤더 불필요:** 클라이언트 코드 레벨에서 JWT나 Bearer 토큰 같은 별도의 HTTP 인증 헤더를 추가할 필요가 없습니다. 인증은 패킷이 Tailscale 터널을 통과할 때 OS/네트워크 레벨에서 자동으로 처리됩니다.
3. **권한 (ACL):** 서버의 `configs/app.json`에 정의된 `allowedUsers` 목록에 클라이언트 기기가 로그인한 Tailscale 계정(이메일)이 포함되어 있어야 합니다.

## 클라이언트 구현 접근 방식

클라이언트(Fyne) 환경에서 Tailscale 네트워크를 타는 방법은 크게 두 가지가 있습니다.

### 접근법 1: OS 레벨 Tailscale 클라이언트 사용 (권장)
가장 쉽고 권장되는 방법입니다.

- **방식:** 사용자의 윈도우, 모바일, 리눅스 기기에 공식 Tailscale 앱을 설치하고 로그인해 둡니다.
- **클라이언트 코드:** Fyne 앱의 코드는 일반적인 HTTP 요청 코드와 100% 동일합니다. OS가 알아서 라우팅을 Tailscale 인터페이스로 넘깁니다.
  ```go
  // OS에 Tailscale이 켜져 있다면, 일반적인 HTTP GET 요청으로 통과됨
  resp, err := http.Get("http://<서버의-Tailscale-IP>:8080/api/v1/monitor")
  ```

### 접근법 2: `tsnet` 패키지를 이용한 내장(Embedded) 모드
사용자 기기에 Tailscale 앱을 설치하기 싫거나, 앱 자체가 단일 Tailscale 노드로 동작하길 원할 때 사용합니다.

- **방식:** Go의 `tailscale.com/tsnet` 패키지를 사용하여 Fyne 앱 내부에 Tailscale 노드를 임베딩합니다.
- **제약:** 최초 실행 시 브라우저를 통한 인증(Auth AuthURL)을 앱 내에서 처리해야 하는 UI 로직이 추가로 필요합니다.
- **클라이언트 코드 예시:**
  ```go
  s := new(tsnet.Server)
  s.Hostname = "fyne-client-app"
  defer s.Close()

  // 일반 http.Client 대신 tsnet이 제공하는 HTTP 클라이언트를 사용
  httpClient := s.HTTPClient()
  resp, err := httpClient.Get("http://<서버의-Tailscale-IP>:8080/api/v1/monitor")
  ```

## 에러 핸들링
클라이언트 UI는 다음 HTTP 상태 코드를 처리해야 합니다.

- **`401 Unauthorized`:** 요청이 Tailscale 터널을 통과하지 않고 일반 인터넷망이나 내부망으로 직접 들어온 경우입니다. (OS의 Tailscale이 꺼져있는지 확인하라는 안내 문구 노출)
- **`403 Forbidden`:** Tailscale 망에는 접속했으나, 해당 계정이 서버의 `allowedUsers`에 등록되지 않은 경우입니다. (관리자 권한이 없다는 안내 문구 노출)
