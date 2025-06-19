-- init.sql
-- Hapus tabel lama jika ada dan ingin rebuild
DROP TABLE IF EXISTS monthly_tax_reports CASCADE;
DROP TABLE IF EXISTS monthly_jobs CASCADE;
DROP TABLE IF EXISTS annual_tax_reports CASCADE;
DROP TABLE IF EXISTS annual_dividend_reports CASCADE;
DROP TABLE IF EXISTS annual_jobs CASCADE;
DROP TABLE IF EXISTS sp2dk_jobs CASCADE;
DROP TABLE IF EXISTS pemeriksaan_jobs CASCADE;
DROP TABLE IF EXISTS invoices CASCADE; -- Tambahkan ini
DROP TABLE IF EXISTS invoice_line_items CASCADE; -- Tambahkan ini
DROP TABLE IF EXISTS clients CASCADE;
DROP TABLE IF EXISTS staffs CASCADE;


CREATE TABLE IF NOT EXISTS clients (
    client_id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_name                 VARCHAR(255) NOT NULL,
    npwp_client                 VARCHAR(20) UNIQUE NOT NULL,
    address_client              TEXT,
    membership_status           VARCHAR(50),
    phone_client                VARCHAR(20),
    email_client                VARCHAR(255),
    pic_client                  VARCHAR(255),
    djp_online_username         VARCHAR(255),
    coretax_username            VARCHAR(255),
    coretax_password_hashed     VARCHAR(255),
    pic_staff_sigma_id          UUID,
    client_category             VARCHAR(100),
    pph_final_umkm              BOOLEAN DEFAULT FALSE,
    pph_25                      BOOLEAN DEFAULT FALSE,
    pph_21                      BOOLEAN DEFAULT FALSE,
    pph_unifikasi               BOOLEAN DEFAULT FALSE,
    ppn                         BOOLEAN DEFAULT FALSE,
    spt_tahunan                 BOOLEAN DEFAULT FALSE,
    pelaporan_deviden           BOOLEAN DEFAULT FALSE,
    laporan_keuangan            BOOLEAN DEFAULT FALSE,
    investasi_deviden           BOOLEAN DEFAULT FALSE,
    created_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS staffs (
    staff_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nip             VARCHAR(20) UNIQUE NOT NULL,
    nama            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hashed VARCHAR(255) NOT NULL,
    role            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE clients
ADD CONSTRAINT fk_pic_staff_sigma
FOREIGN KEY (pic_staff_sigma_id)
REFERENCES staffs (staff_id)
ON DELETE SET NULL;


-- Tabel monthly_jobs (Sudah Benar)
CREATE TABLE IF NOT EXISTS monthly_jobs (
    job_id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id                   UUID NOT NULL,
    job_month                   INT NOT NULL,
    job_year                    INT NOT NULL,
    assigned_pic_staff_sigma_id UUID,
    overall_status              VARCHAR(50) DEFAULT 'Dalam Pengerjaan', -- Standarisasi Default
    proof_of_work_url           TEXT,
    created_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_client_job FOREIGN KEY (client_id) REFERENCES clients (client_id) ON DELETE CASCADE,
    CONSTRAINT fk_assigned_pic_staff_sigma FOREIGN KEY (assigned_pic_staff_sigma_id) REFERENCES staffs (staff_id) ON DELETE SET NULL,
    CONSTRAINT unique_monthly_job_per_client UNIQUE (client_id, job_month, job_year)
);


-- Tabel monthly_tax_reports (Tidak ada perubahan)
CREATE TABLE IF NOT EXISTS monthly_tax_reports (
    report_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id              UUID NOT NULL,
    tax_type            VARCHAR(50) NOT NULL,
    billing_code        VARCHAR(255),
    payment_date        DATE,
    payment_amount      NUMERIC(18, 2),
    report_status       VARCHAR(50) DEFAULT 'Pending',
    report_date         DATE,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_monthly_job FOREIGN KEY (job_id) REFERENCES monthly_jobs (job_id) ON DELETE CASCADE,
    CONSTRAINT unique_tax_report_per_job UNIQUE (job_id, tax_type)
);


-- Tabel annual_jobs (Sudah Benar)
CREATE TABLE IF NOT EXISTS annual_jobs (
    job_id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id                   UUID NOT NULL,
    job_year                    INT NOT NULL,
    assigned_pic_staff_sigma_id UUID,
    overall_status              VARCHAR(50) DEFAULT 'Dalam Pengerjaan', -- Standarisasi Default
    proof_of_work_url           TEXT,
    created_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_client_annual_job FOREIGN KEY (client_id) REFERENCES clients (client_id) ON DELETE CASCADE,
    CONSTRAINT fk_assigned_pic_staff_annual_sigma FOREIGN KEY (assigned_pic_staff_sigma_id) REFERENCES staffs (staff_id) ON DELETE SET NULL,
    CONSTRAINT unique_annual_job_per_client UNIQUE (client_id, job_year)
);

CREATE TABLE IF NOT EXISTS annual_tax_reports (
    report_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL,
    billing_code    VARCHAR(255),
    payment_date    DATE,
    payment_amount  NUMERIC(18, 2),
    report_date     DATE,
    report_status   VARCHAR(50) DEFAULT 'Pending',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_annual_job_tax
        FOREIGN KEY (job_id)
        REFERENCES annual_jobs (job_id)
        ON DELETE CASCADE,
    
    CONSTRAINT unique_annual_tax_report_per_job UNIQUE (job_id)
);


-- Tabel annual_dividend_reports
CREATE TABLE IF NOT EXISTS annual_dividend_reports (
    report_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL,
    is_reported     BOOLEAN DEFAULT FALSE,
    report_date     DATE,
    report_status   VARCHAR(50) DEFAULT 'Pending',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_annual_job_dividend
        FOREIGN KEY (job_id)
        REFERENCES annual_jobs (job_id)
        ON DELETE CASCADE,
    
    CONSTRAINT unique_annual_dividend_report_per_job UNIQUE (job_id)
);


-- Tabel annual_tax_reports, annual_dividend_reports (Tidak ada perubahan)
-- ... (kode Anda untuk tabel ini) ...


-- Tabel sp2dk_jobs <-- PERBAIKAN DI SINI
CREATE TABLE IF NOT EXISTS sp2dk_jobs (
    job_id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id                   UUID NOT NULL,
    assigned_pic_staff_sigma_id UUID,
    contract_no                 VARCHAR(255),
    contract_date               DATE,
    sp2dk_no                    VARCHAR(255),
    sp2dk_date                  DATE,
    bap2dk_no                   VARCHAR(255),
    bap2dk_date                 DATE,
    payment_date                DATE,
    report_date                 DATE,
    overall_status              VARCHAR(50) DEFAULT 'Dalam Pengerjaan', -- <-- DIUBAH
    proof_of_work_url           TEXT,
    created_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sp2dk_client FOREIGN KEY (client_id) REFERENCES clients (client_id) ON DELETE CASCADE,
    CONSTRAINT fk_sp2dk_assigned_pic_staff FOREIGN KEY (assigned_pic_staff_sigma_id) REFERENCES staffs (staff_id) ON DELETE SET NULL
);


-- Tabel pemeriksaan_jobs <-- PERBAIKAN DI SINI
CREATE TABLE IF NOT EXISTS pemeriksaan_jobs (
    job_id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id                   UUID NOT NULL,
    assigned_pic_staff_sigma_id UUID,
    contract_no                 VARCHAR(255),
    contract_date               DATE,
    sp2_no                      VARCHAR(255),
    sp2_date                    DATE,
    skp_no                      VARCHAR(255),
    skp_date                    DATE,
    overall_status              VARCHAR(50) DEFAULT 'Dalam Pengerjaan', -- <-- DIUBAH
    proof_of_work_url           TEXT,
    created_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pemeriksaan_client FOREIGN KEY (client_id) REFERENCES clients (client_id) ON DELETE CASCADE,
    CONSTRAINT fk_pemeriksaan_assigned_pic_staff FOREIGN KEY (assigned_pic_staff_sigma_id) REFERENCES staffs (staff_id) ON DELETE SET NULL
);

-- Tabel invoices dan invoice_line_items (Tidak ada perubahan)
-- ... (kode Anda untuk tabel ini) ...