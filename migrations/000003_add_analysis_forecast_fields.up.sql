-- ADD FORECAST COLUMNS TO ANALYSIS SNAPSHOTS
ALTER TABLE analysis_snapshots ADD COLUMN forecast_end_balance NUMERIC(15,2) DEFAULT 0;
ALTER TABLE analysis_snapshots ADD COLUMN forecast_confidence NUMERIC(4,3) DEFAULT 0;
ALTER TABLE analysis_snapshots ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- UPDATE COMMENT
COMMENT ON COLUMN analysis_snapshots.forecast_end_balance IS 'Prediksi AI untuk saldo di akhir periode';
COMMENT ON COLUMN analysis_snapshots.forecast_confidence IS 'Tingkat kepercayaan AI (0-1) terhadap forecast';
