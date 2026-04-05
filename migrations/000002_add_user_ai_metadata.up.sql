-- ADD AI-FIRST METADATA TO USERS TABLE
ALTER TABLE users ADD COLUMN salary_cycle_day INTEGER NOT NULL DEFAULT 1 CHECK (salary_cycle_day >= 1 AND salary_cycle_day <= 31);
ALTER TABLE users ADD COLUMN financial_goals TEXT[] DEFAULT '{}';
ALTER TABLE users ADD COLUMN risk_profile VARCHAR(20) NOT NULL DEFAULT 'conservative' CHECK (risk_profile IN ('conservative', 'moderate', 'aggressive'));
ALTER TABLE users ADD COLUMN ai_analyst_persona VARCHAR(50) NOT NULL DEFAULT 'supportive';

-- COMMENT ON COLUMNS FOR BETTER DOCS
COMMENT ON COLUMN users.salary_cycle_day IS 'Tanggal gajian user untuk penentuan periode budget & forecast';
COMMENT ON COLUMN users.financial_goals IS 'Daftar objektif finansial user untuk personalisasi saran AI';
COMMENT ON COLUMN users.risk_profile IS 'Profil risiko investasi user';
COMMENT ON COLUMN users.ai_analyst_persona IS 'Karakter suara/gaya bahasa AI saat berinteraksi';
