## Архитектура системы (Микросервисы и gRPC)

```mermaid
flowchart TD
    %% Стиль для клиента
    Client[("👤 Client / Postman")]
    style Client fill:#f9f,stroke:#333,stroke-width:2px

    %% Order Service
    subgraph OS ["📦 Order Service (Port: 8080)"]
        direction TB
        OH["🌐 HTTP Handler"]
        OUC["🧠 Use Case"]
        OR["🗄️ Repository"]
        OGRPC["🔌 gRPC Client"]
    end

    %% Payment Service
    subgraph PS ["💳 Payment Service (Port: 50051)"]
        direction TB
        PH["⚙️ gRPC Handler"]
        PUC["🧠 Use Case"]
        PR["🗄️ Repository"]
    end

    %% Базы данных
    ODB[("🛢️ Order DB")]
    PDB[("🛢️ Payment DB")]

    %% Связи
    Client -->|POST /orders| OH
    Client -->|GET /orders| OH
    
    OH --> OUC
    OUC --> OR
    OUC --> OGRPC
    OR --> ODB

    OGRPC -.->|gRPC Call| PH
    PH --> PUC
    PUC --> PR
    PR --> PDB

    %% Цвета для сервисов
    style OS fill:#e1f5fe,stroke:#01579b
    style PS fill:#f1f8e9,stroke:#33691e