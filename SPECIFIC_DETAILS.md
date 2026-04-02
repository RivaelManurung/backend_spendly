# Specific Implementation Details - Spendly Backend

This document provides a deep dive into the technical implementation of key components in the **Spendly** backend.

## 🧠 AI Analytical Pipeline (`internal/ai` & `internal/service`)
The system uses a **Template-Based AI Engine** to provide financial advice:
- **`GeminiClient`**: A wrapper for Google's Gemini 1.5 Pro. It reads `.prompt` files from `.github/agents/prompts/`, replaces placeholders (e.g., `{{snapshot_data}}`), and enforces JSON responses for structured data extraction.
- **Auto-Categorization**: Uses a specific prompt to classify raw transaction descriptions into one of the system's categories with a confidence score.
- **Workflow Orchestration**: `AnalysisPipeline` executes three AI agents in parallel for the monthly report:
    - `monthly_analysis.prompt`: Summarizes spending behavior.
    - `anomaly_detection.prompt`: Finds unusual spikes compared to previous months.
    - `saving_opportunities.prompt`: Suggests actionable ways to save money.

## 🕒 Automated Background Jobs (`internal/scheduler`)
Scheduled tasks are split into two categories to optimize performance:
1. **Analysis Job (Monthly)**:
    - Aggregates all user transactions for the calendar month.
    - Builds a `CategoryBreakdown` and `DailyTrend` map.
    - Triggers the AI Pipeline to generate insights and stores them in `ai_insights` table.
2. **Budget Monitoring Job (Daily)**:
    - Scans active budgets in the `budgets` table.
    - Compares current spending vs. `limit_amount`.
    - If spending > 80% or 100%, it triggers a real-time AI alert using the `budget_alert.prompt`.

## 🗄️ Database Architecture (`migrations/`)
The schema is designed for scalability with PostgreSQL:
- **`analysis_snapshots`**: Uses **JSONB** columns for `category_breakdown` and `merchant_breakdown`, allowing flexible storage of dynamic data without complex joins.
- **`ai_insights`**: Links directly to a snapshot, preserving the context of why an insight was generated.
- **Optimized Indexing**:
    - `idx_txn_user_date`: B-Tree index for fast transaction history retrieval.
    - `idx_txn_metadata`: GIN index for searching through raw transaction metadata.

## 📡 REST API & Documentation
The API follows a Clean Architecture pattern:
- **DTOs (`internal/handler/dto`)**: Decouple database models from API responses, preventing internal sensitive data leakage (e.g., `deleted_at`).
- **Swagger Integration**: Automatically generated documentation from Go code annotations, allowing frontend developers to test endpoints without reading the source code.
- **Concurrency**: Leverages Go routines for AI generation to ensure the main application remains responsive during heavy analytical tasks.

## 🛠️ Project State (Fixes Applied)
- **Resolved Build Failures**: Fixed multiple `EOF` errors caused by empty files created during initial scaffolding.
- **Import Conflicts**: Resolved name shadowing between `net/http` and `internal/handler/http` by using aliases in `main.go`.
- **Swagger Metadata**: Fixed JSON parsing issues in Swagger by adding `swaggertype` tags to `json.RawMessage` fields.

---
*Created on: 2026-04-01*
