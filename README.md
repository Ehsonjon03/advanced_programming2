## Архитектура проекта (gRPC Interaction)

```mermaid
flowchart TD
    %% Пользователь
    User[("👤 User / Postman")]
    style User fill:#f9f,stroke:#333,stroke-width:2px

    %% Order Service
    subgraph OrderService ["📦 Order Service (HTTP + gRPC Client)"]
        direction TB
        OH["🌐 Gin HTTP Handler"]
        OUC["🧠 Order Use Case"]
        OR["🗄️ Order Repository"]
        OC["🔌 gRPC Client (Generated)"]
    end

    %% Payment Service
    subgraph PaymentService ["💳 Payment Service (gRPC Server)"]
        direction TB
        PH["⚙️ gRPC Handler (Server)"]
        PUC["🧠 Payment Use Case"]
        PR["🗄️ Payment Repository"]
    end

    %% Базы данных
    ODB[("🛢️ order_db")]
    PDB[("🛢️ payment_db")]

    %% Потоки данных
    User -->|REST API: 8080| OH
    OH --> OUC
    OUC --> OR
    OUC --> OC
    OR --> ODB

    %% gRPC Связь
    OC -.->|gRPC Call: 50051| PH
    
    PH --> PUC
    PUC --> PR
    PR --> PDB

    %% Стили
    style OrderService fill:#e3f2fd,stroke:#1565c0
    style PaymentService fill:#fff3e0,stroke:#ef6c00