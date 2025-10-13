# NBOX Architecture Diagram

## Overview

NBOX is a backend service for centralized management of configurations, environment variables, and secrets. It uses hexagonal architecture (ports and adapters) with native AWS integration.

---

## General Architecture

```mermaid
flowchart TB
    subgraph CLIENTS["👥 Clients"]
        CLI["CLI/Scripts"]
        WEB["Web UI"]
        CICD["CI/CD"]
    end

    subgraph PRESENTATION["🌐 Presentation Layer"]
        AUTH["Authentication<br/>JWT + Basic Auth"]
        AUTHZ["Authorization<br/>OPA Policy"]
        API["HTTP API<br/>REST Endpoints"]
        SSE["Server-Sent Events<br/>Real-time Updates"]
    end

    subgraph APPLICATION["⚙️ Application Layer (Use Cases)"]
        ENTRY_UC["EntryUseCase<br/>Variables Management"]
        BOX_UC["BoxUseCase<br/>Templates Management"]
        PATH_UC["PathUseCase<br/>Keys Validation"]
        EVENT_UC["EventUseCase<br/>Notifications"]
        EXPORT_UC["ExportUseCase<br/>Data Export"]
    end

    subgraph DOMAIN["🏛️ Domain"]
        MODELS["Models<br/>Entry | Box | User<br/>Template | Event"]
        PORTS["Interfaces<br/>EntryAdapter<br/>TemplateAdapter<br/>SecretAdapter"]
    end

    subgraph INFRASTRUCTURE["🔌 Adapters"]
        DDB["DynamoDB<br/>Entries + Tracking"]
        S3["S3 Adapter<br/>Templates"]
        SSM["SSM Adapter<br/>Secrets"]
        MEMORY["InMemory<br/>Users"]
        SSE_BROKER["SSE Broker<br/>Events"]
    end

    subgraph AWS["☁️ AWS Services"]
        DYNAMO[("DynamoDB")]
        BUCKET[("S3 Bucket")]
        PARAM[("Parameter Store")]
        KMS[("KMS")]
    end

    CLI --> AUTH
    WEB --> AUTH
    CICD --> AUTH

    AUTH --> AUTHZ
    AUTHZ --> API
    API --> SSE

    API --> ENTRY_UC
    API --> BOX_UC
    API --> EXPORT_UC

    ENTRY_UC --> PATH_UC
    ENTRY_UC --> EVENT_UC
    BOX_UC --> PATH_UC

    ENTRY_UC --> PORTS
    BOX_UC --> PORTS
    EVENT_UC --> PORTS
    EXPORT_UC --> PORTS

    PORTS --> MODELS

    DDB -.implements.-> PORTS
    S3 -.implements.-> PORTS
    SSM -.implements.-> PORTS
    MEMORY -.implements.-> PORTS
    SSE_BROKER -.implements.-> PORTS

    DDB --> DYNAMO
    S3 --> BUCKET
    SSM --> PARAM
    PARAM --> KMS

    SSE --> SSE_BROKER

    classDef clients fill:#E3F2FD,stroke:#1976D2,stroke-width:2px
    classDef presentation fill:#FFF3E0,stroke:#F57C00,stroke-width:2px
    classDef application fill:#E8F5E9,stroke:#388E3C,stroke-width:2px
    classDef domain fill:#FCE4EC,stroke:#C2185B,stroke-width:3px
    classDef infrastructure fill:#F3E5F5,stroke:#7B1FA2,stroke-width:2px
    classDef aws fill:#232F3E,stroke:#FF9900,stroke-width:3px,color:#fff

    class CLI,WEB,CICD clients
    class AUTH,AUTHZ,API,SSE presentation
    class ENTRY_UC,BOX_UC,PATH_UC,EVENT_UC,EXPORT_UC application
    class MODELS,PORTS domain
    class DDB,S3,SSM,MEMORY,SSE_BROKER infrastructure
    class DYNAMO,BUCKET,PARAM,KMS aws
```

---



## Authentication and Authorization Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Auth as Authn/Authz
    participant OPA
    participant UserRepo
    participant Handler

    Client->>API: Request + Credentials
    API->>Auth: Validate Auth

    alt Basic Auth
        Auth->>UserRepo: Check credentials
        UserRepo-->>Auth: User + Roles
    else JWT Token
        Auth->>Auth: Verify JWT signature
        Auth->>Auth: Extract user + roles
    end

    Auth->>OPA: Authorize (user, roles, resource, action)
    OPA-->>Auth: Allow/Deny

    alt Authorized
        Auth-->>API: User context
        API->>Handler: Process request
        Handler-->>Client: Response
    else Denied
        Auth-->>Client: 401/403 Error
    end
```

**Available Roles:**
- `anonymous`: Public health checks
- `viewer`: Read-only access to non-production environments
- `viewer_prod`: Read-only access including production
- `editor`: Read and write access
- `secrets_reader`: Can read plain values of secrets
- `maintainer`: Can delete entries
- `cicd`: Automation access
- `admin`: Full access

---



## Variables Management Flow (Entries)

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant EntryUseCase
    participant SecretAdapter
    participant EntryAdapter
    participant EventUseCase
    participant DynamoDB
    participant SSM as Parameter Store
    participant SSE

    Note over Client,SSE: POST /api/entry (Upsert)

    Client->>API: Upsert entries<br/>[{key, value, secure}]
    API->>EntryUseCase: Upsert(entries)

    EntryUseCase->>EntryUseCase: Separate entries<br/>secure vs normal

    alt Entry is secure
        EntryUseCase->>SecretAdapter: Upsert secrets
        SecretAdapter->>SSM: PutParameter(encrypted)
        SSM-->>SecretAdapter: ARN
        SecretAdapter-->>EntryUseCase: Results with ARN
        EntryUseCase->>EntryUseCase: Replace value with ARN
    end

    EntryUseCase->>EntryAdapter: Upsert entries
    EntryAdapter->>DynamoDB: PutItem (entry)
    EntryAdapter->>DynamoDB: PutItem (tracking)
    DynamoDB-->>EntryAdapter: OK

    EntryAdapter-->>EntryUseCase: Results
    EntryUseCase->>EventUseCase: Emit event
    EventUseCase->>SSE: Broadcast update

    EntryUseCase-->>API: All results
    API-->>Client: Response

    SSE-->>Client: Real-time notification
```

**Available Operations:**
- `POST /api/entry`: Create/update variables
- `GET /api/entry/prefix?v=<path>`: List by prefix
- `GET /api/entry/key?v=<key>`: Get specific variable
- `GET /api/entry/secret-value?v=<key>`: Get plain value of secret
- `DELETE /api/entry/key?v=<key>`: Delete variable
- `GET /api/track/key?v=<key>`: Change history

---



## Templates Management Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant BoxUseCase
    participant S3Adapter
    participant DynamoDB
    participant S3

    Note over Client,S3: POST /api/box (Upsert Template)

    Client->>API: POST /api/box<br/>{service, stage, template}
    API->>BoxUseCase: Upsert template
    BoxUseCase->>BoxUseCase: Decode base64 content

    BoxUseCase->>S3Adapter: Store template
    S3Adapter->>S3: PutObject<br/>(service/stage/template.json)
    S3-->>S3Adapter: Version ID

    S3Adapter->>DynamoDB: Store metadata<br/>(version, timestamp)
    DynamoDB-->>S3Adapter: OK

    S3Adapter-->>BoxUseCase: Success
    BoxUseCase-->>API: Template stored
    API-->>Client: Response

    Note over Client,S3: GET /api/box/{service}/{stage}/{template}/build

    Client->>API: GET template + build<br/>?image-name=nginx:latest
    API->>BoxUseCase: Get and build template
    BoxUseCase->>S3Adapter: Retrieve template
    S3Adapter->>S3: GetObject
    S3-->>S3Adapter: Template content

    BoxUseCase->>BoxUseCase: Process template:<br/>1. Replace :placeholders<br/>2. Replace {{variables}}

    BoxUseCase->>DynamoDB: Retrieve variable values
    DynamoDB-->>BoxUseCase: Values

    BoxUseCase->>BoxUseCase: Build final template
    BoxUseCase-->>API: Processed template
    API-->>Client: Ready-to-use config
```

**Available Operations:**
- `POST /api/box`: Create/update template
- `GET /api/box/{service}/{stage}/{template}`: Get raw template
- `GET /api/box/{service}/{stage}/{template}/build`: Get processed template

**Template Processing:**
1. Placeholders `:variable` - Replaced with query params
2. Variables `{{global/example/var}}` - Replaced with values from DynamoDB/SSM

---



## Export Flow

```mermaid
flowchart LR
    Client["Client"]
    API["API Handler"]
    EXPORT["ExportUseCase"]
    ENTRY["EntryAdapter"]
    FORMAT["Format Exporter"]

    Client -->|"GET /api/export?prefix=dev&format=env"| API
    API --> EXPORT
    EXPORT -->|"List entries"| ENTRY
    ENTRY -->|"Entries"| EXPORT
    EXPORT -->|"Select exporter"| FORMAT

    subgraph EXPORTERS["Exporters"]
        ENV["ENV Exporter<br/>KEY=value"]
        JSON["JSON Exporter<br/>{key: value}"]
        YAML["YAML Exporter<br/>key: value"]
        DOCKER["Docker Compose<br/>environment"]
    end

    FORMAT --> EXPORTERS
    EXPORTERS -->|"Formatted output"| EXPORT
    EXPORT --> API
    API --> Client

    classDef client fill:#E3F2FD,stroke:#1976D2
    classDef usecase fill:#E8F5E9,stroke:#388E3C
    classDef exporter fill:#FFF3E0,stroke:#F57C00

    class Client client
    class EXPORT,ENTRY usecase
    class ENV,JSON,YAML,DOCKER exporter
```

**Supported Formats:**
- `env`: Environment variables (.env)
- `json`: JSON format
- `yaml`: YAML format
- `docker-compose`: For docker-compose.yml

---





## Real-time Event System (SSE)

```mermaid
flowchart TB
    subgraph OPERATIONS["Operations"]
        UPSERT["Upsert Entry"]
        DELETE["Delete Entry"]
        UPDATE["Update Template"]
    end

    subgraph PIPELINE["Event Pipeline"]
        DECORATOR["EntryUseCase<br/>Decorator"]
        EVENT_UC["EventUseCase"]
        BROKER["SSE Event Broker"]
    end

    subgraph CHANNELS["SSE Channels"]
        GLOBAL["Global Channel"]
        PREFIX["Prefix Channels<br/>(dev/, prod/)"]
        SPECIFIC["Specific Key Channels"]
    end

    subgraph CLIENTS["Connected Clients"]
        WEB1["Web UI 1"]
        WEB2["Web UI 2"]
        MONITOR["Monitor Service"]
    end

    UPSERT --> DECORATOR
    DELETE --> DECORATOR
    UPDATE --> DECORATOR

    DECORATOR --> EVENT_UC
    EVENT_UC --> BROKER

    BROKER --> GLOBAL
    BROKER --> PREFIX
    BROKER --> SPECIFIC

    GLOBAL --> WEB1
    GLOBAL --> WEB2
    PREFIX --> WEB1
    SPECIFIC --> MONITOR

    classDef operation fill:#E3F2FD,stroke:#1976D2
    classDef pipeline fill:#E8F5E9,stroke:#388E3C
    classDef channel fill:#FFF3E0,stroke:#F57C00
    classDef client fill:#F3E5F5,stroke:#7B1FA2

    class UPSERT,DELETE,UPDATE operation
    class DECORATOR,EVENT_UC,BROKER pipeline
    class GLOBAL,PREFIX,SPECIFIC channel
    class WEB1,WEB2,MONITOR client
```

**SSE Endpoints:**
- `GET /api/events`: Global stream of all events
- `GET /api/events?prefix=dev`: Filtered stream by prefix

---





## Data Storage

```mermaid
erDiagram
    ENTRIES_TABLE ||--o{ TRACKING_TABLE : tracks
    ENTRIES_TABLE {
        string partition_key
        string value
        bool secure
        string createdAt
        string updatedAt
    }

    TRACKING_TABLE {
        string partition_key
        string sort_key
        string oldValue
        string newValue
        string user
        string action
    }

    BOX_TABLE ||--o{ S3_TEMPLATES : references
    BOX_TABLE {
        string partition_key
        string sort_key
        string s3Key
        string version
        string createdAt
        string updatedAt
    }

    S3_TEMPLATES {
        string key
        string content
        string versionId
    }

    PARAMETER_STORE {
        string name
        string value
        string kmsKeyId
        string type
    }

    ENTRIES_TABLE ||--o{ PARAMETER_STORE : secure_entries_reference
```

---





## Use Cases

### Application Deployment with ECS

```mermaid
sequenceDiagram
    participant DEV as Developer
    participant NBOX
    participant CI as CI/CD Pipeline
    participant ECS
    participant APP as Application

    DEV->>NBOX: 1. Create variables<br/>POST /api/entry<br/>[DB_HOST, API_KEY]
    DEV->>NBOX: 2. Upload task definition template<br/>POST /api/box

    Note over DEV,NBOX: Template contains:<br/>{{dev/myapp/DB_HOST}}<br/>{{dev/myapp/API_KEY}}

    CI->>NBOX: 3. Build template<br/>GET /api/box/myapp/dev/task.json/build
    NBOX-->>CI: Processed task definition

    CI->>ECS: 4. Register task definition
    CI->>ECS: 5. Deploy service

    ECS->>APP: 6. Start containers with<br/>variables and secrets
```



### Secrets Rotation

```mermaid
sequenceDiagram
    participant ADMIN as Admin
    participant NBOX
    participant SSM as Parameter Store
    participant SSE
    participant APPS as Running Apps

    ADMIN->>NBOX: 1. Update secret<br/>POST /api/entry<br/>[{key: "prod/db/password", value: "new-pass", secure: true}]

    NBOX->>SSM: 2. Store encrypted value
    SSM-->>NBOX: ARN

    NBOX->>NBOX: 3. Store ARN reference in DynamoDB
    NBOX->>SSE: 4. Emit event

    SSE-->>APPS: 5. Notify: "prod/db/password updated"

    Note over APPS: Apps can:<br/>- Reload config<br/>- Restart<br/>- Notify status
```

---



## Key Components

### Configuration (Config)
- Loaded from environment variables
- Support for multiple credential strategies
- Configurable environment prefixes

### Path Use Case
- Key validation
- Path normalization
- Allowed prefixes control

### Event System
- Decorator pattern for use cases
- Multi-channel publishing
- SSE support for Web UI

### Health Checks
- AWS connectivity verification
- S3, DynamoDB, SSM status
- Endpoints `/health/ready` and `/health/live`

---



## Security

### Security Layers

1. **Authentication**: HTTP Basic Auth or JWT
2. **Authorization**: OPA (Open Policy Agent) with role-based policies
3. **Encryption**:
   - Secrets encrypted in Parameter Store with KMS
   - TLS in transit (deployment configuration)
4. **Auditing**: Change tracking in DynamoDB table

### OPA Authorization Flow

```mermaid
flowchart LR
    REQUEST["Request"] --> EXTRACT["Extract<br/>user, roles, resource, action"]
    EXTRACT --> OPA["OPA Engine"]
    OPA --> POLICY["policy.rego"]
    OPA --> ROLES["roles.json"]
    OPA --> PERMS["permissions.json"]

    POLICY --> DECISION{Allow?}
    ROLES --> DECISION
    PERMS --> DECISION

    DECISION -->|Yes| ALLOW["Process Request"]
    DECISION -->|No| DENY["403 Forbidden"]

    classDef process fill:#E3F2FD,stroke:#1976D2
    classDef decision fill:#FFF3E0,stroke:#F57C00
    classDef result fill:#C8E6C9,stroke:#2E7D32
    classDef error fill:#FFCDD2,stroke:#C62828

    class REQUEST,EXTRACT,OPA,POLICY,ROLES,PERMS process
    class DECISION decision
    class ALLOW result
    class DENY error
```

---

## Endpoints Summary

### Authentication
- `POST /api/auth/token` - Generate JWT token

### Entries (Variables)
- `POST /api/entry` - Create/update variables
- `GET /api/entry/prefix?v=<path>` - List by prefix
- `GET /api/entry/key?v=<key>` - Get variable
- `GET /api/entry/secret-value?v=<key>` - Get secret (plain)
- `DELETE /api/entry/key?v=<key>` - Delete variable
- `GET /api/track/key?v=<key>` - Change history

### Box (Templates)
- `POST /api/box` - Create/update template
- `GET /api/box/{service}/{stage}/{template}` - Get template
- `GET /api/box/{service}/{stage}/{template}/build` - Get processed template

### Export
- `GET /api/export?prefix=<path>&format=<format>` - Export configuration

### Events (SSE)
- `GET /api/events` - Global event stream
- `GET /api/events?prefix=<path>` - Filtered stream

### Health
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check

---



## Technologies Used

- **Language**: Go 1.24+
- **Framework**: Uber FX (Dependency Injection)
- **HTTP**: Native net/http
- **AWS SDK**: aws-sdk-go-v2
- **Logger**: Zap (structured logging)
- **OPA**: Open Policy Agent for authorization
- **Swagger**: OpenAPI documentation
- **Testing**: Go standard testing

