### 🏗️ Архитектура системы (gRPC взаимодействие)

```mermaid
graph TD
    %% Внешний мир
    Postman[("🚀 Postman / Browser")]
    style Postman fill:#ff6c37,stroke:#333,color:#fff

    subgraph OrderApp ["📦 Order Service (Port 8080)"]
        Gin["🌐 Gin Framework (HTTP)"]
        OrderLogic["🧠 Business Logic"]
        GRPC_Client["🔌 gRPC Client"]
        ListMethod[["📜 Call: ListPayments"]]
    end

    subgraph PaymentApp ["💳 Payment Service (Port 50051)"]
        GRPC_Server["⚙️ gRPC Server"]
        Handler["🛠️ gRPC Handler"]
        UseCase["🧩 UseCase (Logic)"]
        Repo["🗄️ Repo (SQL Filter)"]
    end

    %% Потоки
    Postman -->|HTTP Request| Gin
    Gin --> OrderLogic
    OrderLogic --> GRPC_Client
    GRPC_Client --> ListMethod

    %% gRPC соединение
    ListMethod -.->|Protobuf / gRPC| GRPC_Server

    %% Внутренняя логика Payment
    GRPC_Server --> Handler
    Handler --> UseCase
    UseCase --> Repo

    %% Базы данных
    OrderLogic --- DB1[("🛢️ order_db")]
    Repo --- DB2[("🛢️ payment_db")]

    %% Стилизация
    style OrderApp fill:#e1f5fe,stroke:#01579b
    style PaymentApp fill:#fff3e0,stroke:#ef6c00
    style ListMethod fill:#b3e5fc,stroke:#01579b