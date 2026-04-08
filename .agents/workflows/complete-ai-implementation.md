SPENDLY BACKEND ARCHITECTURE & FEATURE SPECIFICATIONS
Tech Stack: Golang (Echo/Gin), SQLite + GORM, JWT Auth, Google Gemini API

1. CORE AUTHENTICATION & USER MANAGEMENT
   - OAuth2 & JWT Implementation: Secure login/register with refresh token rotation.
   - User Profiling: Secure storage of user preferences (currency, language, dark mode).
   - Biometric Secret Storage: Backend-side validation for biometric-linked sessions.
   - Account Security: Password hashing (Bcrypt) and session management.

2. TRANSACTION ENGINE (The Hub)
   - Multi-Currency Support: Real-time conversion logic for international travel/expense.
   - Advanced CRUD: Support for recurring transactions (subscriptions) with cron-job logic.
   - Transaction Categorization: System-defined and user-defined category mapping.
   - Bulk Operations: Ability to import/export transactions (CSV/JSON).
   - Smart Search: Full-text search using SQLite FTS5 for finding transactions by notes or title.

3. GOAL & BUDGET MANAGEMENT
   - Dynamic Budgeting: Allocation of budgets per category with real-time "Burn Rate" calculation.
   - Goal Progressive Engine: Complex calculation of ETA (Estimated Time of Achievement) for financial goals.
   - Contribution Ledger: Dedicated table for tracking individual contributions to goals.
   - Smart Alerts: Webhook/Notification trigger when spending exceeds 80% of budget.

4. AI FINANCIAL ANALYST PROXY (Secure AI)
   - Proxy Layer: Centralized Gemini API calls to keep API keys hidden from the frontend.
   - Context Builder: Logic to aggregate user finance history into a concise prompt.
   - Insight Caching: Storing AI insights in SQLite to reduce API costs and improve speed.
   - Financial Rating System: An algorithm to give a "Financial Health Score" based on spending habits.

5. SCANNER & OCR MODULE (Image Processing)
   - Receipt Parser: Integration with OCR services (Google Vision/Cloud OCR) to extract amount, date, and merchant.
   - Image Proxying: Secure signed URLs for receipt storage (integrating S3 or local storage).
   - Auto-Categorizer: NLP (Natural Language Processing) to predict categories based on merchant names.

6. DATA SYNC & INTEGRITY
   - Incremental Sync: Logic to sync only updated/new records between mobile and backend.
   - Concurrency Locking: Ensuring data consistency when multiple devices (laptop/mobile) sync at once.
   - Soft Deletes: Implementing "Trash Bin" functionality for recovered deleted transactions.

7. INVOICE GENERATOR (Freelance Module)
   - Proforma & Final Invoicing: Generating dynamic PDF invoices using Go templates.
   - Payment Tracking: Status management for invoices (Draft, Sent, Paid, Overdue).
   - Tax Calculation: Automated tax (VAT/PPN) estimation based on regional rules.

8. ANALYTICS & REPORTING MODULE
   - Monthly Net-Worth: Aggregate calculation of assets vs liabilities.
   - Trend Visualization Data: Optimized JSON endpoints for FL Chart (Mobile-side).
   - Automated Reporting: Scheduled email reports (Weekly/Monthly) containing PDF summaries.

9. ADVANCED SECURITY & DEV OPS
   - Rate Limiting: Preventing API abuse (especially for AI and Auth endpoints).
   - Database Migration: Using `golang-migrate` for structured SQLite schema updates.
   - API Documentation: Automated Swagger/OpenAPI documentation for frontend integration.
   - Health Checks: Endpoint to monitor database and external service status.
10. DATABASE SCHEMA DESIGN (SQLite Optimized)

   A. Table: `users`
      - `id`: TEXT (UUID, Primary Key)
      - `email`: TEXT (Unique, Indexed)
      - `password_hash`: TEXT
      - `name`: TEXT
      - `preferences`: JSON (Stores currency, theme, language)
      - `created_at`: DATETIME

   B. Table: `transactions`
      - `id`: TEXT (Primary Key)
      - `user_id`: TEXT (Foreign Key -> users.id)
      - `title`: TEXT
      - `amount`: REAL
      - `date`: DATETIME (Indexed)
      - `category_id`: TEXT (Foreign Key -> categories.id or goals.id)
      - `type`: TEXT (Check: 'income', 'expense', 'goal')
      - `note`: TEXT
      - `is_recurring`: INTEGER (Boolean 0/1)

   C. Table: `categories`
      - `id`: TEXT (Primary Key)
      - `user_id`: TEXT (Foreign Key -> users.id, Null for system default)
      - `label`: TEXT
      - `icon`: TEXT
      - `color`: TEXT (HEX code)
      - `type`: TEXT ('income', 'expense')

   D. Table: `goals`
      - `id`: TEXT (Primary Key)
      - `user_id`: TEXT (Foreign Key -> users.id)
      - `title`: TEXT
      - `target_amount`: REAL
      - `current_amount`: REAL
      - `icon`: TEXT
      - `color`: TEXT
      - `target_date`: DATETIME
      - `created_at`: DATETIME

   E. Table: `goal_contributions` (Audit Trail for Goals)
      - `id`: TEXT (Primary Key)
      - `goal_id`: TEXT (Foreign Key -> goals.id)
      - `transaction_id`: TEXT (Foreign Key -> transactions.id)
      - `amount`: REAL
      - `date`: DATETIME
      - `note`: TEXT

   F. Table: `budgets`
      - `id`: TEXT (Primary Key)
      - `user_id`: TEXT (Foreign Key -> users.id)
      - `category_id`: TEXT (Foreign Key -> categories.id)
      - `amount`: REAL
      - `period`: TEXT ('monthly', 'weekly')
      - `start_date`: DATETIME

   G. Table: `invoices`
      - `id`: TEXT (Primary Key)
      - `user_id`: TEXT (Foreign Key -> users.id)
      - `client_name`: TEXT
      - `client_email`: TEXT
      - `amount`: REAL
      - `due_date`: DATETIME
      - `status`: TEXT ('draft', 'sent', 'paid', 'overdue')
      - `items`: JSON (Array of {description, price, qty})
      - `created_at`: DATETIME
