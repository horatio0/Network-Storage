# Network Storage 서비스 환경 세팅 가이드

이 문서는 중앙 통제 서버(라즈베리파이 등)와 이를 제어하는 클라이언트(PC) 양쪽에서 이 서비스가 정상 작동하기 위해 필요한 사전 설정들을 설명합니다. 이 서비스는 보안상 **Tailscale** 가상 사설망(VPN) 내부에서만 작동하도록 설계되었습니다.

---

## 1. 서버(Server) 환경 세팅 가이드

중앙에서 파일을 저장하고 백엔드 API를 제공하는 서버(주로 Linux/Raspberry Pi) 설정입니다.

### 1.1 Tailscale 설치 및 설정
모든 통신의 안전망인 Tailscale을 가장 먼저 설치하고 로그인합니다.

```bash
# Tailscale 설치
curl -fsSL https://tailscale.com/install.sh | sh

# Tailscale 실행 및 로그인
sudo tailscale up
```
- 실행 후 터미널에 출력되는 URL에 접속하여 기기를 승인합니다.
- Admin Console에서 서버의 IP(보통 `100.x.x.x` 대역)를 메모해 둡니다. 이는 클라이언트 앱에서 접속할 때 필요합니다.

### 1.2 파일 공유를 위한 마운트 디렉터리 생성
Go 백엔드 서버의 HTTP 파일 업/다운로드 API와 SMB/NFS 서비스가 공통으로 사용할 저장소 디렉터리를 생성합니다.

```bash
# 예시: /home/pi/backend/tmp_mount
mkdir -p /home/pi/backend/tmp_mount
chmod 777 /home/pi/backend/tmp_mount
```
> **주의:** 백엔드의 설정 파일 또는 환경변수 `mountPath` 경로와 이 디렉터리 경로가 반드시 일치해야 합니다. (절대 경로 사용 권장)

### 1.3 Samba (SMB) 세팅 (Windows / Mac 클라이언트 접속용)
클라이언트에서 외부망 노출 없이 Tailscale 인터페이스를 통해서만 익명(Guest)으로 접근 가능하도록 설정합니다.

```bash
sudo apt update
sudo apt install samba -y
```

`sudo nano /etc/samba/smb.conf`를 입력하고, 파일 끝에 아래 내용을 추가합니다.
```ini
[global]
   # Tailscale 인터페이스만 허용
   interfaces = tailscale0
   bind interfaces only = yes
   map to guest = bad user

[NetworkStorage]
   path = /home/pi/backend/tmp_mount
   guest ok = yes
   read only = no
   create mask = 0777
   directory mask = 0777
```

저장 후 Samba 서비스를 재시작합니다.
```bash
sudo systemctl restart smbd
sudo systemctl enable smbd
```

### 1.4 NFS 세팅 (Linux 클라이언트 접속용)
Linux 데스크톱 클라이언트를 위한 NFS 서버 설정입니다.

```bash
sudo apt install nfs-kernel-server -y
```

`sudo nano /etc/exports`를 열고 파일에 아래 내용을 추가합니다. (`100.0.0.0/8`은 Tailscale 고유 IP 대역을 의미합니다.)
```text
/home/pi/backend/tmp_mount  100.0.0.0/8(rw,sync,no_subtree_check,all_squash,anonuid=1000,anongid=1000)
```

저장 후 NFS 서비스를 재시작합니다.
```bash
sudo exportfs -a
sudo systemctl restart nfs-kernel-server
sudo systemctl enable nfs-kernel-server
```

### 1.5 Go 백엔드 서버 빌드 및 구동
백엔드는 제공된 빌드 스크립트를 사용하여 단일 실행 파일로 컴파일할 수 있습니다. 백엔드는 Tailscale의 API를 통해 클라이언트의 신원을 확인하므로, 구동 전 반드시 Tailscale이 실행된 상태여야 합니다.

```bash
cd backend
# 실행 권한 부여 및 빌드 스크립트 실행
chmod +x scripts/build.sh
./scripts/build.sh

# 빌드가 완료되면 bin/ 디렉터리에 backend-server 실행 파일이 생성됩니다.
# 수동으로 직접 실행하려면 아래 명령어를 입력합니다.
./bin/backend-server
```

### 1.6 백엔드를 Systemd 서비스로 등록 (권장)
서버 재부팅 시에도 자동으로 백엔드가 실행되도록 Systemd 데몬 서비스로 등록하는 방법입니다.

1. 서비스 설정 파일 생성
`sudo nano /etc/systemd/system/network-storage-backend.service`를 입력하고 아래 내용을 작성합니다. (경로는 자신의 환경에 맞게 수정하세요)

```ini
[Unit]
Description=Network Storage Backend Server
After=network-online.target tailscaled.service
Wants=network-online.target tailscaled.service

[Service]
Type=simple
# 실제 backend-server 파일이 위치한 절대 경로
ExecStart=/home/pi/backend/bin/backend-server
# 백엔드 서버를 실행할 OS 유저 계정 (예: pi, ubuntu 등)
User=pi
# WorkingDirectory를 backend 폴더의 절대 경로로 지정 (configs/app.json 등을 올바르게 찾기 위해 필요)
WorkingDirectory=/home/pi/backend
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

2. 서비스 데몬 리로드 및 활성화
```bash
sudo systemctl daemon-reload
sudo systemctl enable network-storage-backend
sudo systemctl start network-storage-backend
```

3. 서비스 상태 및 로그 확인
```bash
# 상태 확인
sudo systemctl status network-storage-backend
# 실시간 로그 확인
journalctl -u network-storage-backend -f
```

---

## 2. 클라이언트(Client) 환경 세팅 가이드

서버에 접속하여 파일 관리, 터미널 제어, 및 시스템 모니터링을 수행할 데스크톱 앱(Fyne Frontend) 설정입니다.

### 2.1 Tailscale 설치 및 로그인
클라이언트 PC(Windows, Mac, Linux)에도 Tailscale을 설치해야 합니다.
- [Tailscale 다운로드 페이지](https://tailscale.com/download)에서 OS에 맞는 클라이언트를 설치합니다.
- 설치 후 서버를 등록했던 **동일한 Tailscale 계정(Tailnet)**으로 로그인합니다.

### 2.2 Fyne 프론트엔드 구동 환경 준비
Go 언어 기반의 데스크톱 UI 라이브러리인 Fyne을 실행하기 위한 개발 환경 요구사항입니다.
- **Go 설치**: 최소 Go 1.20 이상의 버전을 설치합니다.
- **C 컴파일러 (CGO 요구사항)**: 데스크톱 UI 렌더링을 위해 필요합니다.
  - **Windows**: MSYS2, TDM-GCC, 또는 MSVC 설치가 필요합니다.
  - **Mac**: Xcode Command Line Tools가 필요합니다. (`xcode-select --install` 실행)
  - **Linux**: 빌드 필수 패키지 설치 (`sudo apt install gcc libgl1-mesa-dev xorg-dev` 등)

### 2.3 프론트엔드 앱 실행
필수 조건이 충족되었다면, 클라이언트 코드 디렉터리로 이동하여 앱을 실행합니다.

```bash
cd fyne-frontend
go mod tidy
go run main.go
```
- 앱이 실행되면 화면 우측의 **설정(Settings)** 메뉴로 이동합니다.
- **Server IP** 칸에 `1.1` 단계에서 확인한 **서버의 Tailscale IP (100.x.x.x)**를 입력합니다.
- **Server Port**는 백엔드에서 설정된 포트(기본: `8080`)를 입력하고 저장합니다.
- 설정 완료 후, 대시보드(Dashboard)와 파일 관리자(Files) 등에서 데이터가 정상적으로 불러와지는지 통신 상태를 확인합니다.

### 2.4 (선택) OS별 수동 네트워크 드라이브 연결
Fyne 프론트엔드 앱 내부의 파일 탐색기를 사용하지 않고 OS 자체의 파일 탐색기(Finder, 윈도우 탐색기 등)를 통해 직접 마운트하고 싶다면 아래 명령어를 참고하세요. (`<Tailscale_IP>` 자리에 서버의 100.x.x.x IP 입력)

- **Windows:** 
  `net use Z: \\<Tailscale_IP>\NetworkStorage` 
  *(또는 탐색기의 '네트워크 드라이브 연결' 메뉴 사용)*
- **Mac:** 
  `mount_smbfs //guest:@<Tailscale_IP>/NetworkStorage /Volumes/NetworkStorage` 
  *(또는 Finder에서 '서버로 연결' `smb://<Tailscale_IP>/NetworkStorage`)*
- **Linux:** 
  `sudo mount -t nfs <Tailscale_IP>:/home/pi/backend/tmp_mount /mnt/NetworkStorage`
