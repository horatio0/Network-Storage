# Files API

파일 업로드, 다운로드 및 디렉터리 목록 조회를 위한 API.

## 1. 파일 목록 조회

지정한 경로(`path`)의 파일 및 폴더 목록을 조회합니다.

- **URL**: `/api/v1/files/list`
- **Method**: `GET`
- **Query Params**:
  - `path`: 탐색할 상대 경로 (예: `/`, `/subfolder`). Path Traversal 방어를 위해 마운트된 디렉터리 외부로 나가는 경로는 차단됩니다.
- **Response**:
  - **Status 200 OK**:
    ```json
    [
      {
        "name": "example.txt",
        "isDir": false,
        "size": 1024,
        "modTime": "2026-06-19T12:00:00Z"
      },
      {
        "name": "subfolder",
        "isDir": true,
        "size": 4096,
        "modTime": "2026-06-19T12:01:00Z"
      }
    ]
    ```
  - **Status 404 Not Found**: 경로를 찾을 수 없음
  - **Status 403 Forbidden**: 유효하지 않은 경로 접근 시도

## 2. 파일 업로드

파일을 마운트된 디렉터리에 업로드합니다.

- **URL**: `/api/v1/files/upload`
- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **Form Data**:
  - `file`: 업로드할 파일 객체
- **Response**:
  - **Status 200 OK**:
    ```json
    {
      "message": "file uploaded successfully",
      "filename": "uploaded.txt"
    }
    ```

## 3. 파일 다운로드

지정한 파일을 다운로드합니다.

- **URL**: `/api/v1/files/download`
- **Method**: `GET`
- **Query Params**:
  - `filename`: 다운로드할 파일명 (단일 파일 이름만 허용. 디렉터리 이름 X)
- **Response**:
  - **Status 200 OK**: 파일 바이너리 데이터
  - **Status 404 Not Found**: 파일을 찾을 수 없음

## 4. 디렉터리 생성

마운트된 디렉터리 내부에 새로운 디렉터리를 생성합니다. 부모 디렉터리가 없다면 함께 생성됩니다.

- **URL**: `/api/v1/files/mkdir`
- **Method**: `POST`
- **Query Params**:
  - `path`: 생성할 디렉터리의 상대 경로 (예: `/new_folder`, `/sub/new_folder`)
- **Response**:
  - **Status 200 OK**:
    ```json
    {
      "message": "directory created successfully"
    }
    ```
  - **Status 400 Bad Request**: 경로 파라미터 누락
  - **Status 403 Forbidden**: 마운트된 디렉터리 외부로 나가는 경로 차단
  - **Status 500 Internal Server Error**: 디렉터리 생성 실패

## 5. 파일 및 디렉터리 삭제

마운트된 디렉터리 내부의 파일 또는 디렉터리를 재귀적으로 완전히 삭제합니다. 최상위 마운트 경로 자체는 삭제할 수 없습니다.

- **URL**: `/api/v1/files/delete`
- **Method**: `DELETE`
- **Query Params**:
  - `path`: 삭제할 대상의 상대 경로 (예: `/example.txt`, `/subfolder`)
- **Response**:
  - **Status 200 OK**:
    ```json
    {
      "message": "deleted successfully"
    }
    ```
  - **Status 400 Bad Request**: 경로 파라미터 누락
  - **Status 403 Forbidden**: 마운트된 디렉터리 외부 접근 또는 최상위 마운트 디렉터리 삭제 시도
  - **Status 500 Internal Server Error**: 삭제 실패
