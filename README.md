graph TD
%% Внешний мир
Postman[("🚀 Postman / Browser")]
style Postman fill:#ff6c37,stroke:#333,color:#fff,rx:10,ry:10

    subgraph OrderApp ["📦 Order Service (Port 8080)"]
        Gin["🌐 Gin Framework (HTTP)"]
        OrderLogic["🧠 Business Logic"]
        GRPC_Client["🔌 gRPC Client"]
        
        %% Изменение 1: Покажем, что клиент вызывает
        ListMethod[["📜 New: Call ListPayments"]]
        style ListMethod fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    end

    subgraph PaymentApp ["💳 Payment Service (Port 50051)"]
        GRPC_Server["⚙️ gRPC Server"]
        
        %% Изменение 2: Покажем новые слои реализации
        Handler["🛠️ gRPC Handler"]
        UseCase["🧩 UseCase (List Logic)"]
        Repo["🗄️ Repo (SQL Filter)"]
        
        style Handler fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
        style UseCase fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
        style Repo fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    end

    %% Потоки
    Postman -->|HTTP Request /orders| Gin
    Gin --> OrderLogic
    OrderLogic -->|Call| GRPC_Client
    GRPC_Client --> ListMethod

    %% Тот самый gRPC переход
    ListMethod -.->|Protobuf / gRPC / 50051| GRPC_Server

    %% Изменение 3: Внутренний поток Payment Service
    GRPC_Server --> Handler
    Handler -->|1. GetStatus()| UseCase
    UseCase -->|2. Filter| Repo

    %% Базы
    OrderLogic --- DB1[("🛢️ order_db")]
    Repo --- DB2[("🛢️ payment_db")]
    
    %% Показываем, что данные возвращаются
    Repo -.->|Filtered Data| DB2
    Repo ==>|3. Domain Models| UseCase
    UseCase ==>|4. Proto Messages| Handler
    Handler ==>|5. ListPaymentsResponse| GRPC_Server

    %% Стилизация
    style OrderApp fill:#e1f5fe,stroke:#01579b,rx:15,ry:15
    style PaymentApp fill:#fff3e0,stroke:#ef6c00,rx:15,ry:15
    linkStyle 4 stroke:#ef6c00,stroke-width:2px,stroke-dasharray: 5 5;
    linkStyle 8,9,10,11 stroke:#ef6c00,stroke-width:3px;