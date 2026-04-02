# Money Tracker AI Agent

Go backend + Antigravity agent workflow untuk aplikasi money tracker dengan AI analyst.

## Stack
- **Backend**: Go 1.22
- **Database**: PostgreSQL 16
- **AI**: Gemini 1.5 Pro (primary analyst) + GPT-4o Mini (categorizer)
- **Agent runtime**: Antigravity
- **Push**: Firebase Cloud Messaging (FCM)

---

## Project structure

```
money-tracker-agent/
├── agents/
│   ├── money_tracker.agent          # Main Antigravity workflow definition
│   └── prompts/
│       ├── categorize.prompt        # Auto-categorization (GPT-4o-mini)
│       ├── ocr_parse.prompt         # Receipt OCR extraction
│       ├── monthly_analysis.prompt  # Monthly AI report
│       ├── anomaly_detection.prompt # Spending anomaly detector
│       ├── saving_opportunities.prompt
│       ├── chat_analyst.prompt      # Free-text Q&A
│       ├── intent_parser.prompt     # Parse user chat intent
│       ├── daily_digest.prompt      # Daily spending summary
│       └── budget_alert.prompt      # Budget threshold notification
    Readme
│
├── internal/agent
│   ├── transaction/
│   │   ├── validate.go       # Input validation + currency normalization
│   │   ├── create.go         # Single + bulk transaction creation
│   │   └── transform_csv.go  # BCA / Mandiri / BRI CSV parser
│   ├── category/
│   │   └── auto_categorize.go # Rule-based + LLM categorization
│   ├── analysis/
│   │   ├── build_snapshot.go  # Aggregate transactions → snapshot
│   │   └── chat_analyst.go    # Chat Q&A with user's financial data
│   ├── budget/
│   │   └── check_threshold.go # Budget 80/90/100% alert trigger
│   ├── ai/
│   │   └── save_insight.go    # Store + retrieve AI insights
│   ├── ocr/
│   │   └── extract.go         # Receipt image → structured data
│   └── report/
│       └── generate.go        # CSV / PDF export
│
├── pkg/
│   ├── llm/
│   │   └── client.go   # Unified Gemini + OpenAI client
│   ├── db/
│   │   └── db.go       # PostgreSQL connection pool
│   └── notify/
│       └── push.go     # FCM push + internal event emitter
│
├── migrations/
│   └── 001_init.sql    # Full DB schema + seed data
│
├── go.mod
├── .env.example
└── README.md
```

---

## Setup

```bash
# 1. Clone & install deps
go mod download

# 2. Config
cp .env.example .env
# Fill in GEMINI_API_KEY, OPENAI_API_KEY, DATABASE_URL, etc.

# 3. Database
psql $DATABASE_URL -f migrations/001_init.sql

# 4. Register agent with Antigravity
antigravity deploy agents/money_tracker.agent

# 5. Run locally
go run main.go
```

---

## Agent workflows & triggers

| Workflow | Trigger | Frequency |
|---|---|---|
| `transaction_input_pipeline` | `POST /transactions` | On demand |
| `ocr_scan_pipeline` | `POST /transactions/ocr` | On demand |
| `csv_import_pipeline` | `POST /transactions/import` | On demand |
| `auto_categorize_pipeline` | event `transaction.created` | Real-time |
| `budget_alert_pipeline` | event `budget.threshold_reached` | Real-time |
| `daily_insight_job` | cron `0 8 * * *` | Daily 08:00 |
| `monthly_snapshot_job` | cron `0 0 1 * *` | 1st of month |
| `chat_analyst_pipeline` | event `user.message` | Real-time |
| `report_export_pipeline` | `GET /reports/export` | On demand |

---

## AI pipeline flow

```
Transaction created
    │
    ├─► rule-based match (free, fast, ~0ms)
    │       confidence >= 0.90 → save suggestion ✓
    │
    └─► LLM classify (GPT-4o-mini, ~300ms)
            → save suggestion with confidence score
            → user confirms/rejects in UI

Monthly (1st of month)
    │
    ├─► build_snapshot (aggregate raw txns)
    ├─► monthly_analysis    (Gemini 1.5 Pro)
    ├─► anomaly_detection   (Gemini 1.5 Pro)
    └─► saving_opportunities (Gemini 1.5 Pro)
            → save all insights to ai_insights table
            → push notification to user
```

---

## Key decisions

**Why `decimal` not `float64` for money?**  
Float precision errors compound. `shopspring/decimal` does exact arithmetic — `0.1 + 0.2 = 0.3`, not `0.30000000000000004`.

**Why rule-based first, then LLM?**  
~70% of transactions (Gojek, Indomaret, Netflix) can be categorized with zero API cost via merchant rules. LLM is only called for ambiguous cases.

**Why `analysis_snapshots`?**  
Pre-aggregation means AI analyst never scans raw transaction rows. One snapshot row = one month of data. Fast, cheap, and consistent inputs to the LLM.

**Why two AI models?**  
GPT-4o-mini: fast + cheap for high-frequency categorization.  
Gemini 1.5 Pro: long context (1M tokens) for monthly analysis with full breakdown data.
