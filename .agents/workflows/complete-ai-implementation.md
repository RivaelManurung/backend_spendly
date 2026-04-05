---
description: Rencana Implementasi Akhir Spendly AI-First (User Personality + Cash Flow Forecast)
---

# 🤖 Workflow: Finalisasi Spendly AI-First

Workflow ini akan menjalankan sisa implementasi teknis untuk menyempurnakan sistem AI pada Spendly Backend.

## 🚀 Langkah 1: Update API Interface User (Personality Data)
Agar data gajian dan goals bisa diisi oleh user, kita perlu mengupdate handler dan service.

// turbo
1. Update `internal/service/user_service.go` untuk menangani field `SalaryCycleDay` dan `FinancialGoals`.
2. Update HTTP Request/Response model di layer handler (misal `pkg/dto/user.custom.go` atau file sejenis).

## 🧠 Langkah 2: Perkuat Prompt Engineering & AI Analis
Kita akan menyatukan data **Forecast** dengan analisa rutin bulanan ke dalam AI Insight.

1. Hubungkan `ForecastEndBalance` ke dalam variabel prompt di `internal/service/analysis_pipeline.go`.
2. Tes kaitan antara `SalaryCycleDay` user dengan perhitungan sisa hari dalam periode analisis.

## 📊 Langkah 3: Finalisasi Fitur Deteksi Anomali & Langganan
Mengaktifkan agent khusus untuk mencari bocor-bocor biaya.

1. Implementasi fungsi pencarian pola transaksi berulang (Subscription Radar).
2. Integrasi alert anomali pengeluaran tinggi mendadak.

---

> [!NOTE]
> Setelah workflow ini dibuat, saya akan mulai menjalankan **Langkah 1** secara otomatis.
