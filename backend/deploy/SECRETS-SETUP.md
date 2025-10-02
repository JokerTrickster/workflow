# GitHub Secrets Setup Guide

## Required Secrets for Backend Deployment

Navigate to: **GitHub Repository → Settings → Secrets and variables → Actions → New repository secret**

### 1. EC2_CLOUD_HOST
- **Name**: `EC2_CLOUD_HOST`
- **Value**: Your EC2 public IP address
- **Example**: `13.203.37.93`

### 2. EC2_CLOUD_USER
- **Name**: `EC2_CLOUD_USER`
- **Value**: SSH username for EC2
- **Example**: `ubuntu`

### 3. EC2_LOCAL_HOST (Optional - for local environment)
- **Name**: `EC2_LOCAL_HOST`
- **Value**: Your local EC2 IP address
- **Example**: `192.168.1.100`

### 4. EC2_LOCAL_USER (Optional - for local environment)
- **Name**: `EC2_LOCAL_USER`
- **Value**: SSH username for local EC2
- **Example**: `ubuntu`

### 5. EC2_SSH_KEY
- **Name**: `EC2_SSH_KEY`
- **Value**: Your private SSH key content

#### How to get SSH private key:

```bash
# On your local machine, display the private key
cat ~/.ssh/id_rsa

# Or if you use a specific key for EC2
cat ~/.ssh/your-ec2-key.pem
```

Copy the **entire key** including:
- `-----BEGIN RSA PRIVATE KEY-----`
- All the key content
- `-----END RSA PRIVATE KEY-----`

**Example format**:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdefghijklmnopqrstuvwxyz...
(many more lines)
...xyz9876543210
-----END RSA PRIVATE KEY-----
```

## EC2 Security Group Configuration

Ensure your EC2 security group allows:

### Inbound Rules:
- **SSH (22)**: From your IP or GitHub Actions IP ranges
- **HTTP (7000)**: From S3 website or 0.0.0.0/0 (if public API)
- **MySQL (3306)**: From EC2 instance IP (internal)
- **RabbitMQ (5672)**: From EC2 instance IP (internal)

### Example Security Group:
```
Type        Protocol    Port    Source
SSH         TCP         22      0.0.0.0/0 (or your IP)
Custom TCP  TCP         7000    0.0.0.0/0 (API access)
MySQL       TCP         3306    sg-xxxxx (internal only)
RabbitMQ    TCP         5672    sg-xxxxx (internal only)
```

## Verification Checklist

Before deploying, verify:

- [ ] All 5 secrets are configured in GitHub
- [ ] SSH key has correct format (includes BEGIN/END lines)
- [ ] EC2 security group allows SSH (port 22)
- [ ] EC2 security group allows API traffic (port 7000)
- [ ] SSH key matches the one configured on EC2
- [ ] EC2 instance is running
- [ ] You can manually SSH to EC2: `ssh -i ~/.ssh/key.pem ubuntu@13.203.37.93`

## Testing Secrets

Test if secrets are accessible in GitHub Actions:

1. Create a test workflow:

```yaml
name: Test Secrets
on: workflow_dispatch

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Check secrets
        env:
          EC2_HOST: ${{ secrets.EC2_CLOUD_HOST }}
          EC2_USER: ${{ secrets.EC2_CLOUD_USER }}
        run: |
          echo "Host: ${EC2_HOST:0:5}..." # Show first 5 chars only
          echo "User: $EC2_USER"
          echo "SSH Key exists: ${{ secrets.EC2_SSH_KEY != '' }}"
```

2. Run the workflow
3. Check if secrets are resolved (partial values shown for security)

## Common Issues

### Issue: "Permission denied (publickey)"
**Solution**:
- Verify SSH key format includes BEGIN/END lines
- Check EC2 key pair matches the secret
- Ensure key has correct permissions: `chmod 600 ~/.ssh/key.pem`

### Issue: "Host key verification failed"
**Solution**:
- The workflow automatically adds host to known_hosts
- Check if EC2_CLOUD_HOST IP is correct

### Issue: Secrets not found
**Solution**:
- Verify secret names are exactly: `EC2_CLOUD_HOST`, `EC2_CLOUD_USER`, `EC2_SSH_KEY`
- Check secrets are in "Actions" section, not "Codespaces" or "Dependabot"

## Security Best Practices

1. **Never commit secrets to git**
   - Use `.gitignore` for `.env` files
   - Use GitHub Secrets for CI/CD

2. **Rotate secrets regularly**
   - Update SSH keys every 90 days
   - Update API tokens when compromised

3. **Limit secret access**
   - Only give access to necessary workflows
   - Use environment-specific secrets

4. **Audit secret usage**
   - Review workflow runs
   - Check for unauthorized access

## Quick Setup Commands

```bash
# 1. Generate new SSH key (if needed)
ssh-keygen -t rsa -b 4096 -C "github-actions@workflow"

# 2. Copy public key to EC2
ssh-copy-id -i ~/.ssh/id_rsa.pub ubuntu@13.203.37.93

# 3. Test SSH connection
ssh -i ~/.ssh/id_rsa ubuntu@13.203.37.93

# 4. Display private key for GitHub Secret
cat ~/.ssh/id_rsa
```

Copy the output and add as `EC2_SSH_KEY` secret.
