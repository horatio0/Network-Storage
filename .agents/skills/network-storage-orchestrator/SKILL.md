---
name: network-storage-orchestrator
description: "Network Storage (Go Backend + Fyne Frontend) 프로젝트의 기능을 구현, 수정, 검증할 때 사용하는 메인 워크플로우 오케스트레이터. '기능 추가해줘', 'UI 수정해줘', 'API 연동해줘', '버그 고쳐줘', '다시 실행해줘', '보완해줘' 등의 요청 시 반드시 이 스킬을 트리거할 것."
---

# Network Storage Orchestrator Workflow

Network Storage 프로젝트의 변경사항을 처리하기 위해 에이전트 팀(Backend, Frontend, QA)을 조율하는 오케스트레이터 스킬.

**실행 모드:** 에이전트 팀 (Agent Team)

## 워크플로우

### Phase 0: 컨텍스트 확인
1. 사용자의 요청이 신규 기능 추가인지, 기존 워크스페이스의 부분 재실행/수정인지 판별한다.
2. 이전 작업 내역(예: `_workspace/` 폴더 내의 산출물)이 있으면 이를 읽고 반영한다.

### Phase 1: 요구사항 및 영향도 분석
1. 사용자의 요청(프롬프트)을 분석하여 백엔드(`backend/`)와 프론트엔드(`fyne-frontend/`)에 미치는 영향을 파악한다.
2. 작업 목록(Task List)을 작성한다.

### Phase 2: 팀 구성 및 작업 할당 (invoke_subagent)
1. `invoke_subagent` 도구를 사용하여 에이전트 팀을 구성한다.
   - **Backend Developer**: `subagent_type: general-purpose`, 모델: `기본 모델 (Gemini 3.1 Pro)`. 역할: 백엔드 로직 및 API 구현 (규칙은 `.agents/backend_dev.md` 참조)
   - **Frontend Developer**: `subagent_type: general-purpose`, 모델: `기본 모델 (Gemini 3.1 Pro)`. 역할: Fyne UI 및 API 연동 (규칙은 `.agents/frontend_dev.md` 참조)
   - **QA Tester**: `subagent_type: general-purpose`, 모델: `기본 모델 (Gemini 3.1 Pro)`. 역할: 프론트-백엔드 경계면 검증 및 통합 테스트 (규칙은 `.agents/qa_tester.md` 참조)
2. 각 에이전트에게 구체적인 작업 지시(Prompt)를 할당하여 스폰한다.

### Phase 3: 실행 및 조율 (send_message)
1. 에이전트들은 `send_message` 도구를 사용하여 스스로 API 스펙이나 요구사항을 조율한다.
2. 오케스트레이터는 진행 상황을 모니터링하고 중간 산출물을 추적한다.

### Phase 4: 검증 (QA)
1. 백엔드와 프론트엔드 작업이 완료되면 QA 에이전트가 경계면 교차 검증을 수행한다.
2. 결함이 발견되면 QA 에이전트가 담당 에이전트에게 수정을 요청(`send_message`)한다.

### Phase 5: 최종 보고
1. 모든 검증이 통과되면 오케스트레이터는 작업 결과를 종합하여 사용자에게 최종 보고한다.
2. 사용자에게 피드백이 있는지 질문하고 세션을 종료한다.

## 데이터 전달 프로토콜
- **메시지 기반**: `send_message`를 통해 에이전트 간 실시간 API 스펙 합의.
- **파일 기반**: 대규모 작업 명세서나 중간 산출물은 `_workspace/` 하위에 파일로 저장하여 공유.

## 에러 핸들링
- 컴파일 에러나 테스트 실패 시 에이전트들은 2회까지 자체 수정을 시도하며, 해결되지 않을 경우 오케스트레이터가 개입하여 사용자에게 상황을 보고하고 판단을 요청한다.
