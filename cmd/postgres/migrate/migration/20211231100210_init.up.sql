-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS "balance"
(
    "accountId"      INT8   NOT NULL PRIMARY KEY,
    "balance"        float4 NOT NULL,
    "depositAllSum"  float4 NOT NULL CHECK ( "depositAllSum" >= 0),
    "depositCount"   INT    NOT NULL,
    "pincoinBalance" float4 NOT NULL CHECK ( "pincoinBalance" >= 0 ),
    "pincoinAllSum"  float4 NOT NULL CHECK ( "pincoinAllSum" >= 0 )
);

CREATE TABLE IF NOT EXISTS "journal"
(
    "id"              UUID     NOT NULL, -- _id
    "id2"             bytea       NOT NULL,               -- id
    "accountId"       INT8        NOT NULL,
    "created_at"      TIMESTAMP   NOT NULL,
    "balance"         FLOAT8   DEFAULT NULL,
    "depositAllSum"   FLOAT8   DEFAULT NULL,
    "depositCount"    INT      DEFAULT NULL,
    "pincoinBalance"  FLOAT8   DEFAULT NULL,
    "pincoinAllSum"   FLOAT8   DEFAULT NULL,
    "change"          FLOAT4   DEFAULT NULL,
    "pincoinChange"   FLOAT4   DEFAULT NULL,
    "currency"        SMALLINT DEFAULT NULL,
    "project"         VARCHAR(64) NOT NULL,
    "revert"          BOOLEAN  DEFAULT NULL,
    "transactionId"   INT8        NOT NULL,
    "transactionBson" bytea       NOT NULL,
    "transactionType" VARCHAR(36) NOT NULL
) PARTITION BY RANGE (created_at);

CREATE INDEX IF NOT EXISTS journal_id_idx ON journal (id);
CREATE INDEX IF NOT EXISTS journal_created_at_idx ON journal (created_at);
CREATE TABLE IF NOT EXISTS journal_default PARTITION OF journal DEFAULT;

-- создаем нужные партиции для данных на основе данных + 60 месяцев вперед (5 лет!)
DO
$$
    DECLARE
        REC RECORD;
        cmd text;
    BEGIN
        FOR REC IN
            WITH dates AS (
                SELECT current_timestamp                 as min_date
                     , date_trunc('month', CURRENT_DATE) as max_date
            ),
                 part AS (
                     SELECT to_char(g.month, 'YYYYMM')    as mm
                          , g.month                       as mm_beg
                          , g.month + '1 month'::interval as mm_end
                     FROM dates,
                          generate_series(dates.min_date, dates.max_date + '60 month'::interval,
                                          '1 month'::interval) AS g(month)
                 )
            SELECT mm, mm_beg, mm_end
            FROM part
            LOOP
                -- create partition of month
                cmd = format('CREATE TABLE IF NOT EXISTS %s_%s PARTITION OF %s FOR VALUES FROM (%L) TO (%L);',
                             'journal',
                             REC.mm, 'journal', REC.mm_beg, REC.mm_end);
                --RAISE notice '%', cmd;
                EXECUTE cmd;
            END LOOP;
    END
$$;
