-- Create "user" table
CREATE TABLE "user" (
  "id" bigint NOT NULL,
  "first_name" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create "running_man_library" table
CREATE TABLE "running_man_library" (
  "id" bigint NOT NULL,
  "year" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "running_man_library_year_key" UNIQUE ("year")
);
-- Create "running_man_video" table
CREATE TABLE "running_man_video" (
  "id" uuid NOT NULL,
  "running_man_library_year" integer NOT NULL,
  "episode" integer NOT NULL,
  "price" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "running_man_video_episode_key" UNIQUE ("episode"),
  CONSTRAINT "fk_video_library" FOREIGN KEY ("running_man_library_year") REFERENCES "running_man_library" ("year") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "collection" table
CREATE TABLE "collection" (
  "user_id" bigint NOT NULL,
  "running_man_video_episode" integer NOT NULL,
  PRIMARY KEY ("user_id", "running_man_video_episode"),
  CONSTRAINT "fk_user_collection" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_video_collection" FOREIGN KEY ("running_man_video_episode") REFERENCES "running_man_video" ("episode") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "transaction" table
CREATE TABLE "transaction" (
  "id" uuid NOT NULL,
  "user_id" bigint NOT NULL,
  "running_man_video_episode" integer NOT NULL,
  "amount" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_transaction" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_video_transaction" FOREIGN KEY ("running_man_video_episode") REFERENCES "running_man_video" ("episode") ON UPDATE CASCADE ON DELETE CASCADE
);
