-- Modify "user" table
ALTER TABLE "user" ADD COLUMN "subscription_expired_at" timestamptz NULL, ADD COLUMN "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, ADD COLUMN "deleted_at" timestamptz NULL;
-- Rename a column from "running_man_video_episode" to "total_amount"
ALTER TABLE "invoice" RENAME COLUMN "running_man_video_episode" TO "total_amount";
-- Modify "invoice" table
ALTER TABLE "invoice" DROP CONSTRAINT "fk_user_invoice", DROP CONSTRAINT "fk_video_invoice", ADD CONSTRAINT "invoice_total_amount_check" CHECK (total_amount >= 0), DROP COLUMN "amount", ADD COLUMN "ref_id" character varying(255) NOT NULL, ADD COLUMN "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_invoice_user_id" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Modify "payment" table
ALTER TABLE "payment" DROP CONSTRAINT "fk_invoice_payment", DROP CONSTRAINT "fk_user_payment", ADD CONSTRAINT "payment_amount_paid_check" CHECK (amount_paid >= 0), ADD COLUMN "amount_paid" integer NOT NULL, ADD COLUMN "status" character varying(50) NOT NULL, ADD COLUMN "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_payment_invoice_id" FOREIGN KEY ("invoice_id") REFERENCES "invoice" ("id") ON UPDATE CASCADE ON DELETE CASCADE, ADD
 CONSTRAINT "fk_payment_user_id" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Create "library" table
CREATE TABLE "library" (
  "id" bigint NOT NULL,
  "year" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "library_year_key" UNIQUE ("year")
);
-- Create "video" table
CREATE TABLE "video" (
  "id" uuid NOT NULL,
  "library_year" integer NOT NULL,
  "episode" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "video_episode_key" UNIQUE ("episode"),
  CONSTRAINT "fk_video_library_year" FOREIGN KEY ("library_year") REFERENCES "library" ("year") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Drop "collection" table
DROP TABLE "collection";
-- Drop "running_man_video" table
DROP TABLE "running_man_video";
-- Drop "running_man_library" table
DROP TABLE "running_man_library";
