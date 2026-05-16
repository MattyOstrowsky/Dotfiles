---
name: vault-secrets
description: Secrets management patterns for HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager, Kubernetes Sealed Secrets, and External Secrets Operator
---

## General Principles

### Never Hardcode Secrets
- No secrets in: source code, CI/CD logs, environment variables, ConfigMaps, Dockerfiles
- All secrets in a dedicated secret store
- Access via IAM/RBAC — never share credentials

### Secret Types
| Type | Store | Rotation |
|------|-------|----------|
| Database passwords | Vault / AWS SM | 90 days auto |
| API tokens | Vault / AWS SM | On compromise |
| TLS certificates | cert-manager / ACM | 90 days auto |
| SSH keys | Vault SSH CA | Per session |
| Encryption keys | KMS / Cloud KMS | Annual |

## HashiCorp Vault

### Static Secrets
```bash
# Write
vault kv put secret/myapp/database \
  username=db_user \
  password=$(openssl rand -base64 32)

# Read
vault kv get secret/myapp/database
vault kv get -field=password secret/myapp/database

# Delete
vault kv delete secret/myapp/database

# List versions
vault kv list secret/myapp/
```

### Dynamic Secrets (Database)
```bash
# Enable database secrets engine
vault secrets enable database

# Configure PostgreSQL
vault write database/config/postgres \
  plugin_name=postgresql-database-plugin \
  allowed_roles="myapp-role" \
  connection_url="postgresql://{{username}}:{{password}}@host:5432/db" \
  username="vault_admin" \
  password="secret"

# Create role (short-lived credentials)
vault write database/roles/myapp-role \
  db_name=postgres \
  creation_statements="CREATE USER \"{{name}}\" WITH PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

# Get credentials
vault read database/creds/myapp-role
```

### Vault Agent Sidecar (Kubernetes)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      serviceAccountName: myapp
      containers:
      - name: myapp
        image: myapp:1.0.0
        volumeMounts:
        - name: secrets
          mountPath: /etc/secrets
          readOnly: true
      - name: vault-agent
        image: hashicorp/vault:1.16
        args:
        - agent
        - -config=/etc/vault/config.hcl
        volumeMounts:
        - name: vault-config
          mountPath: /etc/vault
        - name: secrets
          mountPath: /etc/secrets
      volumes:
      - name: vault-config
        configMap:
          name: vault-agent-config
      - name: secrets
        emptyDir:
          medium: Memory
```

## AWS Secrets Manager

### Terraform
```hcl
resource "aws_secretsmanager_secret" "db_password" {
  name                    = "prod/myapp/database"
  recovery_window_in_days = 7
  kms_key_id             = aws_kms_key.secrets.arn
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = jsonencode({
    username = "appuser"
    password = random_password.db.result
    host     = aws_db_instance.main.address
    port     = 5432
  })
}
```

### Application Access
```python
import boto3
import json
from botocore.config import Config

def get_secret(secret_name: str, region: str = "eu-central-1") -> dict:
    """Retrieve secret from AWS Secrets Manager."""
    client = boto3.client(
        "secretsmanager",
        region_name=region,
        config=Config(connect_timeout=5, read_timeout=5)
    )
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response["SecretString"])
```

## Kubernetes Secrets Management

### Sealed Secrets
```bash
# Encrypt a secret for K8s (safe to commit to Git)
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml

# Decrypt (requires controller private key)
kubeseal --recovery-unseal --recovery-private-key sealed-secrets-key.yaml < sealed-secret.yaml

# Original secret.yaml (DO NOT commit this)
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secret
data:
  api-key: <base64>

# Sealed secret.yaml (SAFE to commit)
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: myapp-secret
spec:
  encryptedData:
    api-key: <encrypted-blob>
```

### External Secrets Operator
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: eu-central-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: database-secret       # Kubernetes Secret name
    creationPolicy: Owner
  data:
  - secretKey: DATABASE_URL
    remoteRef:
      key: prod/myapp/database
      property: connection_string
```

## Rotation Strategies

### Automated Rotation (AWS Lambda example)
```python
import boto3
import json
import secrets
import string

def lambda_handler(event, context):
    """Rotate database password in Secrets Manager."""
    secret_id = event["SecretId"]
    client = boto3.client("secretsmanager")
    
    # Generate new password
    alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
    new_password = "".join(secrets.choice(alphabet) for _ in range(32))
    
    # Update secret
    client.put_secret_value(
        SecretId=secret_id,
        SecretString=json.dumps({
            "username": "appuser",
            "password": new_password
        })
    )
    
    # Update database password (implementation depends on DB type)
    update_database_password("appuser", new_password)
    
    return {"status": "rotated"}
```

## Anti-Patterns
- ❌ `.env` files in version control
- ❌ Secrets in CI/CD pipeline logs (no `echo $SECRET`)
- ❌ Long-lived static credentials for services
- ❌ Same secret across environments (dev/staging/prod)
- ❌ No rotation policy defined
- ❌ Using default encryption keys
- ❌ Access logs disabled on secret store
