-- Create "invoice" table
CREATE TABLE "invoice" (
  "id" uuid NOT NULL,
  "user_id" bigint NOT NULL,
  "running_man_video_episode" integer NOT NULL,
  "amount" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "expired_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_invoice" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_video_invoice" FOREIGN KEY ("running_man_video_episode") REFERENCES "running_man_video" ("episode") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "payment" table
CREATE TABLE "payment" (
  "id" character varying(255) NOT NULL,
  "user_id" bigint NOT NULL,
  "invoice_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "payment_invoice_id_key" UNIQUE ("invoice_id"),
  CONSTRAINT "fk_invoice_payment" FOREIGN KEY ("invoice_id") REFERENCES "invoice" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_user_payment" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Drop "transaction" table
DROP TABLE "transaction";
