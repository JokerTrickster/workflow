# SSH Key 설정 가이드

## 문제 상황
```
ubuntu@13.203.37.93: Permission denied (publickey)
```
이 오류는 SSH 키가 EC2 인스턴스에 등록되지 않았음을 의미합니다.

## 해결 방법

### 방법 1: 기존 EC2 키페어 사용 (권장)

EC2 인스턴스 생성 시 다운로드한 `.pem` 파일이 있다면:

```bash
# 1. .pem 파일 권한 설정
chmod 400 ~/Downloads/your-key.pem

# 2. SSH 접속 테스트
ssh -i ~/Downloads/your-key.pem ubuntu@13.203.37.93

# 3. 접속 성공하면, 이 키를 GitHub Secrets에 추가
cat ~/Downloads/your-key.pem
```

출력된 전체 내용을 복사해서 GitHub Secrets의 `EC2_SSH_KEY`에 추가합니다.

### 방법 2: 새 SSH 키 생성 및 등록

#### Step 1: SSH 키 생성

```bash
# 1. 새 SSH 키 생성
ssh-keygen -t rsa -b 4096 -f ~/.ssh/ec2-workflow-key -C "github-actions@workflow"

# Enter 입력 (비밀번호 없이)
# Enter 입력 (확인)

# 2. 생성 확인
ls -la ~/.ssh/ec2-workflow-key*
# ec2-workflow-key      (개인키 - GitHub Secret용)
# ec2-workflow-key.pub  (공개키 - EC2에 등록)
```

#### Step 2: EC2에 공개키 등록

**옵션 A: AWS Console 사용 (기존 .pem 파일 있을 때)**

```bash
# 1. 기존 .pem으로 접속
ssh -i ~/Downloads/your-key.pem ubuntu@13.203.37.93

# 2. EC2에서 authorized_keys에 새 공개키 추가
echo "본인_로컬에서_복사한_공개키_내용" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# 3. 로컬에서 공개키 내용 복사
cat ~/.ssh/ec2-workflow-key.pub
```

**옵션 B: AWS Systems Manager Session Manager 사용**

AWS Console → EC2 → 인스턴스 선택 → Connect → Session Manager

```bash
# Session Manager로 접속 후
sudo su - ubuntu
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# 공개키 추가
echo "여기에_공개키_붙여넣기" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

**옵션 C: AWS EC2 User Data 스크립트**

AWS Console → EC2 → 인스턴스 → Stop → Actions → Instance Settings → Edit user data

```bash
#!/bin/bash
echo "여기에_공개키_붙여넣기" >> /home/ubuntu/.ssh/authorized_keys
chmod 600 /home/ubuntu/.ssh/authorized_keys
chown ubuntu:ubuntu /home/ubuntu/.ssh/authorized_keys
```

인스턴스 시작 → 다시 Stop → User Data 제거

#### Step 3: 접속 테스트

```bash
# 새 키로 접속 테스트
ssh -i ~/.ssh/ec2-workflow-key ubuntu@13.203.37.93

# 성공하면 ✅
```

#### Step 4: GitHub Secrets 등록

```bash
# 개인키 전체 내용 복사
cat ~/.ssh/ec2-workflow-key

# 출력 예시:
# -----BEGIN OPENSSH PRIVATE KEY-----
# b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdz...
# ...
# -----END OPENSSH PRIVATE KEY-----
```

GitHub → Repository → Settings → Secrets → Actions → New secret
- Name: `EC2_SSH_KEY`
- Value: (위에서 복사한 전체 내용)

### 방법 3: 로컬 SSH Config 설정 (로컬 개발용)

```bash
# ~/.ssh/config 파일 생성/수정
nano ~/.ssh/config
```

다음 내용 추가:

```
Host workflow-ec2
    HostName 13.203.37.93
    User ubuntu
    IdentityFile ~/.ssh/ec2-workflow-key
    StrictHostKeyChecking no
```

이제 간단하게 접속:

```bash
ssh workflow-ec2
```

## 검증 체크리스트

### ✅ EC2 측
```bash
# SSH로 EC2 접속 후
ls -la ~/.ssh/
# drwx------ (700)  .ssh/
# -rw------- (600)  authorized_keys

cat ~/.ssh/authorized_keys
# ssh-rsa AAAAB3NzaC1yc2EA... 로 시작하는 공개키 확인
```

### ✅ 로컬 측
```bash
# 개인키 권한 확인
ls -la ~/.ssh/ec2-workflow-key
# -rw------- (600)

# 개인키와 공개키 매칭 확인
ssh-keygen -y -f ~/.ssh/ec2-workflow-key > /tmp/test.pub
diff /tmp/test.pub ~/.ssh/ec2-workflow-key.pub
# 차이 없으면 OK
```

### ✅ GitHub Secrets
```yaml
# .github/workflows/test-ssh.yml
name: Test SSH Connection
on: workflow_dispatch

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test SSH
        env:
          EC2_HOST: ${{ secrets.EC2_CLOUD_HOST }}
          EC2_USER: ${{ secrets.EC2_CLOUD_USER }}
          EC2_SSH_KEY: ${{ secrets.EC2_SSH_KEY }}
        run: |
          mkdir -p ~/.ssh
          echo "$EC2_SSH_KEY" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          ssh-keyscan -H $EC2_HOST >> ~/.ssh/known_hosts
          ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no $EC2_USER@$EC2_HOST 'echo "SSH connection successful!"'
```

## 문제 해결

### 1. "Permission denied" 여전히 발생
```bash
# EC2에서 SSH 데몬 로그 확인
sudo tail -f /var/log/auth.log

# 일반적인 원인:
# - 공개키가 authorized_keys에 없음
# - authorized_keys 권한 문제 (600이 아님)
# - .ssh 디렉토리 권한 문제 (700이 아님)
```

### 2. "Host key verification failed"
```bash
# 로컬에서 known_hosts 정리
ssh-keygen -R 13.203.37.93

# 다시 접속
ssh -i ~/.ssh/ec2-workflow-key ubuntu@13.203.37.93
```

### 3. GitHub Actions에서만 실패
```bash
# Secret 값 확인 (민감정보 마스킹됨)
- name: Debug
  run: |
    echo "Key length: ${#EC2_SSH_KEY}"
    echo "First 20 chars: ${EC2_SSH_KEY:0:20}..."
  env:
    EC2_SSH_KEY: ${{ secrets.EC2_SSH_KEY }}
```

## 보안 권장사항

1. **개인키 보호**
   ```bash
   # 반드시 600 권한
   chmod 600 ~/.ssh/ec2-workflow-key

   # Git에 절대 커밋하지 않기
   echo "*.pem" >> .gitignore
   echo "*_key" >> .gitignore
   ```

2. **공개키만 공유**
   ```bash
   # .pub 파일만 공유 가능
   cat ~/.ssh/ec2-workflow-key.pub
   ```

3. **정기적인 키 로테이션**
   - 90일마다 새 키 생성
   - 기존 키 제거

4. **EC2 보안 그룹**
   - SSH(22) 포트는 특정 IP만 허용 권장
   - GitHub Actions IP 범위 추가 필요시:
     - https://api.github.com/meta 확인

## 빠른 설정 스크립트

```bash
#!/bin/bash
# setup-ssh-key.sh

echo "🔑 SSH Key Setup for EC2"

# 1. 키 생성
ssh-keygen -t rsa -b 4096 -f ~/.ssh/ec2-workflow-key -N ""

# 2. 공개키 출력
echo ""
echo "📋 Copy this PUBLIC key to EC2 ~/.ssh/authorized_keys:"
echo "----------------------------------------"
cat ~/.ssh/ec2-workflow-key.pub
echo "----------------------------------------"

# 3. 개인키 출력 (GitHub Secret용)
echo ""
echo "🔐 Copy this PRIVATE key to GitHub Secret (EC2_SSH_KEY):"
echo "----------------------------------------"
cat ~/.ssh/ec2-workflow-key
echo "----------------------------------------"

# 4. SSH Config 추가
echo ""
echo "📝 Adding to SSH config..."
cat >> ~/.ssh/config << EOF

Host workflow-ec2
    HostName 13.203.37.93
    User ubuntu
    IdentityFile ~/.ssh/ec2-workflow-key
    StrictHostKeyChecking no
EOF

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Add public key to EC2: ssh-copy-id -i ~/.ssh/ec2-workflow-key.pub ubuntu@13.203.37.93"
echo "2. Or manually: ssh ubuntu@13.203.37.93 (using existing key) and paste public key"
echo "3. Test: ssh workflow-ec2"
echo "4. Add private key to GitHub Secrets as EC2_SSH_KEY"
```

사용법:
```bash
chmod +x setup-ssh-key.sh
./setup-ssh-key.sh
```
