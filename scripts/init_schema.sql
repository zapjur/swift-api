CREATE TABLE IF NOT EXISTS swift_codes (
    swift_code VARCHAR(11) PRIMARY KEY,
    bank_name TEXT NOT NULL,
    address TEXT,
    town_name TEXT NOT NULL,
    country_iso2 CHAR(2) NOT NULL,
    country_name TEXT NOT NULL,
    timezone TEXT NOT NULL,
    is_headquarter BOOLEAN NOT NULL,
    headquarter_swift_code VARCHAR(11),

    FOREIGN KEY (headquarter_swift_code) REFERENCES swift_codes(swift_code)
    );

CREATE INDEX IF NOT EXISTS idx_country_iso2 ON swift_codes(country_iso2);
CREATE INDEX IF NOT EXISTS idx_headquarter_swift ON swift_codes(headquarter_swift_code);
