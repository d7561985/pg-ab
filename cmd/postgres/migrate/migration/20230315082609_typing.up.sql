ALTER TABLE balance
    ALTER COLUMN balance TYPE bigint,
    ALTER COLUMN "pincoinBalance" TYPE bigint,
    ALTER COLUMN "depositAllSum" TYPE bigint,
    ALTER COLUMN "pincoinAllSum"  TYPE bigint;


ALTER TABLE journal
    ALTER COLUMN balance TYPE bigint,
    ALTER COLUMN "pincoinBalance" TYPE bigint,
    ALTER COLUMN "change" TYPE INT,
    ALTER COLUMN "pincoinChange"  TYPE INT;