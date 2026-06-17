# File Transfer API Guide

이 문서는 Phase 3에서 구현된 파일 전송 API의 클라이언트 연동 가이드입니다. 
서버는 설정된 `mountPath` (기본값: `tmp_mount`) 디렉터리 내에서만 파일 작업을 허용하여 경로 이탈(Path Traversal) 공격을 방지합니다.

## 공통 사항
- **Base URL:** `http://<server-ip>:8080` (또는 설정된 포트)
- **인증:** 모든 API는 Phase 1에서 구현된 Tailscale 네트워크 인증을 통과해야 합니다. 클라이언트는 Tailscale 네트워크 위에서 API를 호출해야 합니다.

---

## 1. 파일 업로드 API

서버의 마운트 디렉터리에 파일을 업로드합니다.

- **URL:** `/api/v1/files/upload`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`

### Request

`file` 필드에 업로드할 파일 데이터를 담아 전송해야 합니다.

**cURL 예시:**
```bash
curl -X POST http://localhost:8080/api/v1/files/upload \
  -F "file=@/path/to/local/test.txt"
```

### Response

- **200 OK:** 성공
  ```json
  {
    "message": "file uploaded successfully",
    "filename": "test.txt"
  }
  ```
- **400 Bad Request:** 파일 필드가 없거나 잘못된 파일명(경로 탈출 시도 등)인 경우.
- **500 Internal Server Error:** 서버 측 저장 실패.

---

## 2. 파일 다운로드 API

서버의 마운트 디렉터리에 있는 파일을 다운로드합니다.

- **URL:** `/api/v1/files/download`
- **Method:** `GET`

### Request Parameters

| 파라미터 | 타입 | 필수 여부 | 설명 |
| :--- | :--- | :--- | :--- |
| `filename` | string | 필수 | 다운로드할 파일의 이름 |

**cURL 예시:**
```bash
curl -O -J http://localhost:8080/api/v1/files/download?filename=test.txt
```

### Response

- **200 OK:** 파일 스트림 반환 (`Content-Type: application/octet-stream` 등)
- **400 Bad Request:** `filename` 쿼리 파라미터가 누락되었거나 유효하지 않은 경우.
- **404 Not Found:** 요청한 파일이 존재하지 않는 경우.

### 보안 참고 (Path Traversal 방지)
클라이언트가 `filename=../../../etc/passwd`와 같이 상위 디렉터리 접근을 시도하더라도, 서버는 안전하게 파일의 기본 이름(`passwd`)만 추출하여 `tmp_mount` 내부에서만 탐색합니다. 따라서 클라이언트는 파일명만 정확히 전달하면 됩니다.
