#!/usr/bin/env bash

set -euo pipefail

mode="${1:-fast}"

case "$mode" in
  fast|full)
    ;;
  *)
    echo "usage: $0 [fast|full]" >&2
    exit 2
    ;;
esac

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-/tmp/go-build-cache}"
mkdir -p "$GOCACHE"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

info() {
  echo "==> $*"
}

check_format() {
  local -a go_files=()
  local path

  mapfile -t all_approved < <(approved_paths)
  for path in "${all_approved[@]}"; do
    if [[ "$path" == *.go ]] && [[ -f "$path" ]]; then
      go_files+=("$path")
    fi
  done

  if ((${#go_files[@]} == 0)); then
    info "format: no approved Go files to check"
    return
  fi

  mapfile -t dirty < <(gofmt -l "${go_files[@]}")
  if ((${#dirty[@]} > 0)); then
    printf 'FAIL: gofmt mismatch:\n' >&2
    printf '  %s\n' "${dirty[@]}" >&2
    fail "run gofmt before completing task"
  fi

  info "format: ok"
}

package_list() {
  local pkg
  while IFS= read -r pkg; do
    case "$mode" in
      fast)
        # Add packages to skip in fast mode if any
        ;;
    esac
    printf '%s\n' "$pkg"
  done < <(go list ./...)
}

run_tests() {
  mapfile -t packages < <(package_list)
  if ((${#packages[@]} == 0)); then
    fail "no packages selected for test run"
  fi

  info "test: running $mode suite"
  go test "${packages[@]}"
}

approved_paths() {
  local task_file
  task_file="$(resolve_task_file)" || return 0
  perl -ne '
    if (/Approved-Files:\s*(.+)$/) {
      my @items = split /,/, $1;
      for my $item (@items) {
        $item =~ s/^\s+|\s+$//g;
        print "$item\n" if $item ne "none";
      }
    }
  ' "$task_file"
}

resolve_task_file() {
  local candidate=""
  local latest=""
  local latest_mtime=0

  if [[ -n "${TASK_FILE:-}" ]]; then
    [[ -f "$TASK_FILE" ]] || fail "TASK_FILE does not exist: $TASK_FILE"
    printf '%s\n' "$TASK_FILE"
    return
  fi

  shopt -s nullglob
  for candidate in plan/tasks/*.md; do
    local mtime
    mtime="$(stat -c '%Y' "$candidate")"
    if ((mtime > latest_mtime)); then
      latest="$candidate"
      latest_mtime="$mtime"
    fi
  done
  shopt -u nullglob

  [[ -n "$latest" ]] || return 1
  printf '%s\n' "$latest"
}

doc_acknowledged() {
  local task_file
  task_file="$(resolve_task_file)" || return 1
  grep -Eiq 'Docs-Impact:[[:space:]]*none' "$task_file" &&
    grep -Eiq 'Docs-Reason:[[:space:]]*.+$' "$task_file"
}

require_task_metadata() {
  local task_file="$1"
  grep -Eiq 'Requirements-Clarity:[[:space:]]*(clear|unclear)' "$task_file" ||
    fail "task file must declare Requirements-Clarity"
  grep -Eiq 'Clarification-Status:[[:space:]]*(resolved|pending)' "$task_file" ||
    fail "task file must declare Clarification-Status"
  grep -Eiq 'Assumptions-Used:[[:space:]]*(no|yes)' "$task_file" ||
    fail "task file must declare Assumptions-Used"
  grep -Eiq 'Assumption-Approval:[[:space:]]*(not-required|approved|pending)' "$task_file" ||
    fail "task file must declare Assumption-Approval"
  grep -Eiq 'Function-Length-Exception:[[:space:]]*(no|yes)' "$task_file" ||
    fail "task file must declare Function-Length-Exception"
  grep -Eiq 'Function-Length-Approval:[[:space:]]*(not-required|approved|pending)' "$task_file" ||
    fail "task file must declare Function-Length-Approval"
  grep -Eiq 'Implementation-Plan-Status:[[:space:]]*(drafted|pending|approved)' "$task_file" ||
    fail "task file must declare Implementation-Plan-Status"
  grep -Eiq 'Implementation-Step-Status:[[:space:]]*(planning-only|awaiting-user-approval|approved-for-implementation)' "$task_file" ||
    fail "task file must declare Implementation-Step-Status"
  grep -Eiq 'Implementation-Approval:[[:space:]]*(approved|pending)' "$task_file" ||
    fail "task file must declare Implementation-Approval"
}

check_requirement_governance() {
  local task_file
  local clarity_status
  local assumption_status

  task_file="$(resolve_task_file)" || fail "missing task file under plan/tasks/; create one from docs/harness/task-template.md"
  require_task_metadata "$task_file"

  if grep -Eiq 'Requirements-Clarity:[[:space:]]*unclear' "$task_file"; then
    grep -Eiq 'Clarification-Status:[[:space:]]*resolved' "$task_file" ||
      fail "unclear requirements require Clarification-Status: resolved before implementation"
    grep -Eiq 'Clarification-Questions:[[:space:]]*(.+|none)' "$task_file" ||
      fail "unclear requirements must record Clarification-Questions"
    grep -Eiq 'Clarification-Answer:[[:space:]]*(.+|none)' "$task_file" ||
      fail "unclear requirements must record Clarification-Answer"
  fi

  if grep -Eiq 'Assumptions-Used:[[:space:]]*yes' "$task_file"; then
    grep -Eiq 'Assumption-Approval:[[:space:]]*approved' "$task_file" ||
      fail "assumptions require user approval before implementation"
    grep -Eiq 'Approval-Evidence:[[:space:]]*.+$' "$task_file" ||
      fail "approved assumptions must record Approval-Evidence"
  fi

  if grep -Eiq 'Function-Length-Exception:[[:space:]]*yes' "$task_file"; then
    grep -Eiq 'Function-Length-Approval:[[:space:]]*approved' "$task_file" ||
      fail "function length exceptions require user approval before implementation"
    grep -Eiq 'Function-Length-Evidence:[[:space:]]*.+$' "$task_file" ||
      fail "function length exceptions must record Function-Length-Evidence"
  fi

  if grep -Eiq 'Implementation-Plan-Status:[[:space:]]*(drafted|pending)' "$task_file"; then
    fail "implementation plan must be approved before writing code"
  fi

  if grep -Eiq 'Implementation-Step-Status:[[:space:]]*(planning-only|awaiting-user-approval)' "$task_file"; then
    fail "implementation step is still waiting for approval"
  fi

  grep -Eiq 'Implementation-Approval:[[:space:]]*approved' "$task_file" ||
    fail "implementation plan requires user approval before implementation"
  grep -Eiq 'Implementation-Approval-Evidence:[[:space:]]*.+$' "$task_file" ||
    fail "approved implementation plans must record Implementation-Approval-Evidence"

  clarity_status="$(grep -Eio 'Requirements-Clarity:[[:space:]]*(clear|unclear)' "$task_file" | tail -n 1)"
  assumption_status="$(grep -Eio 'Assumptions-Used:[[:space:]]*(no|yes)' "$task_file" | tail -n 1)"
  info "task governance: ok (${clarity_status:-unknown}, ${assumption_status:-unknown})"
}

function_length_exception_allowed() {
  local task_file
  task_file="$(resolve_task_file)" || return 1
  grep -Eiq 'Function-Length-Exception:[[:space:]]*yes' "$task_file" &&
    grep -Eiq 'Function-Length-Approval:[[:space:]]*approved' "$task_file"
}



check_function_length() {
  local -a go_files=()
  local output=""
  local path

  mapfile -t all_approved < <(approved_paths)
  for path in "${all_approved[@]}"; do
    if [[ "$path" == *.go ]] && [[ -f "$path" ]]; then
      go_files+=("$path")
    fi
  done

  if ((${#go_files[@]} == 0)); then
    info "length: no approved Go files to check"
    return
  fi

  if function_length_exception_allowed; then
    info "length: approved exception found in $(resolve_task_file)"
    return
  fi

  output="$(
    perl -ne '
      our ($in_func, $sig, $start, $depth, $line);
      $line = $.;
      if (!$in_func && /^\s*func\b/) {
        $in_func = 1;
        $sig = $_;
        $start = $line;
        $depth = 0;
      }
      if ($in_func) {
        $depth += tr/{/{/;
        $depth -= tr/}/}/;
        if ($depth == 0 && /\}/) {
          my $len = $line - $start + 1;
          if ($len > 15) {
            chomp($sig);
            print "$ARGV:$start:$len:$sig\n";
          }
          $in_func = 0;
          $sig = q{};
        }
      }
    ' "${go_files[@]}"
  )"

  if [[ -n "$output" ]]; then
    printf 'FAIL: function or method length exceeded 15 lines:\n' >&2
    while IFS= read -r line; do
      printf '  %s\n' "$line" >&2
    done <<< "$output"
    fail "split large functions or record an approved exception in the task file"
  fi

  info "length: ok"
}

check_doc_sync() {
  local -a approved=()
  local need_architecture=0
  local path

  mapfile -t approved < <(approved_paths)
  if ((${#approved[@]} == 0)); then
    info "docs: no approved changes, skipping doc sync gate"
    return
  fi

  for path in "${approved[@]}"; do
    case "$path" in
      docs/architecture/architecture.ko.md|docs/conventions/directory-convention.ko.md|docs/conventions/type-reference.ko.md)
        ;;
      internal/ui/*|internal/client/*|internal/app/*|internal/config/*|main.go|configs/app.json)
        need_architecture=1
        ;;
    esac
  done

  if ((need_architecture == 0)); then
    info "docs: no synced docs required for approved paths"
    return
  fi

  if doc_acknowledged; then
    info "docs: task waiver found in $(resolve_task_file)"
    return
  fi

  if ((need_architecture == 1)) && ! printf '%s\n' "${approved[@]}" | grep -Eq '^(docs/architecture/architecture\.ko\.md|docs/conventions/directory-convention\.ko\.md)$'; then
    fail "approved structural changes require docs/architecture/architecture.ko.md or docs/conventions/directory-convention.ko.md"
  fi

  info "docs: ok"
}

check_requirement_governance

check_function_length
check_format
run_tests
check_doc_sync

info "agent check completed"
