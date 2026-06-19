# WebRTC 시그널링 API 가이드

이 문서는 `central-control-backend`에서 제공하는 WebRTC 초기 연결 협상(Signaling)을 위한 WebSocket API 연동 방법을 설명한다.

## 개요
WebRTC를 통한 P2P 화면 공유 및 원격 제어를 위해서는 두 기기 간에 세션 정보(Offer, Answer)와 네트워크 정보(ICE Candidate)를 교환해야 한다. 이 API는 그 정보를 상대방에게 1:1로 다이렉트 중계(Relay)해주는 허브 역할을 한다.

## 엔드포인트
- **URL:** `ws://<server-ip>:<port>/api/v1/signaling/ws`
- **프로토콜:** WebSocket (JSON Message)
- **보안:** Tailscale 네트워크망 접근 및 인증 필요

## 식별자 (ID)
이 시그널링 서버는 별도의 세션 ID를 발급하지 않고, **Tailscale 노드 이름(`ts_node`)**을 고유 식별자로 사용한다.
따라서 내가 메시지를 보낼 상대방(Target)의 Tailscale 기기 이름을 알아야 한다.

## 메시지 규격
클라이언트가 서버로 보내고, 서버가 타겟 클라이언트에게 그대로 전달하는 JSON 규격이다.

```json
{
  "type": "offer", 
  "sender": "my-laptop", // 서버가 자동으로 삽입/보정하므로 전송 시 생략해도 됨
  "target": "raspberrypi-5", // 메시지를 받을 상대방의 ts_node 이름 (필수)
  "payload": { ... } // WebRTC의 RTCSessionDescription 또는 RTCIceCandidate 원본 객체
}
```

- `type`: 메시지의 종류 (`offer`, `answer`, `candidate` 등 자율 지정)
- `sender`: 메시지를 보낸 사람의 `ts_node`. (위조 방지를 위해 서버가 연결된 세션의 실제 ID로 덮어씌워서 타겟에게 전달함)
- `target`: 메시지를 받을 사람의 `ts_node`. 이 값이 허브에 연결되어 있어야 전달됨.
- `payload`: 실제 WebRTC 협상 데이터. 서버는 이 내용을 파싱하지 않고 그대로 타겟에게 전달(Bypass)함.

## 통신 흐름 예시 (A -> B 연결)
1. 기기 A와 기기 B가 모두 `ws://.../api/v1/signaling/ws`에 접속.
2. 기기 A가 WebRTC `Offer`를 생성하여 JSON 포맷으로 서버에 전송 (target: "B").
3. 서버가 B에게 해당 JSON 전달 (sender: "A"가 자동으로 붙음).
4. 기기 B가 `Offer`를 받고 `Answer`를 생성하여 서버에 전송 (target: "A").
5. 서버가 A에게 해당 JSON 전달.
6. A와 B가 `ICE Candidate`들을 위와 같은 방식으로 계속 교환.
7. WebRTC P2P 연결 성립 성공.
