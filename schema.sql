CREATE TABLE "user" (
  id BIGINT PRIMARY KEY,
  first_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE running_man_library (
  id BIGINT PRIMARY KEY,
  year INT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE running_man_video (
  id UUID PRIMARY KEY,
  running_man_library_year INT NOT NULL,
  episode INT UNIQUE NOT NULL,
  price INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  CONSTRAINT fk_video_library
    FOREIGN KEY (running_man_library_year)
    REFERENCES running_man_library (year)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);

CREATE TABLE collection (
  user_id BIGINT NOT NULL,
  running_man_video_episode INT NOT NULL,

  PRIMARY KEY (user_id, running_man_video_episode),

  CONSTRAINT fk_user_collection
    FOREIGN KEY (user_id)
    REFERENCES "user" (id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT fk_video_collection
    FOREIGN KEY (running_man_video_episode)
    REFERENCES running_man_video (episode)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);

CREATE TABLE invoice (
  id UUID PRIMARY KEY,
  user_id BIGINT NOT NULL,
  running_man_video_episode INT NOT NULL,
  amount INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  expired_at TIMESTAMPTZ NOT NULL,

  CONSTRAINT fk_user_invoice
    FOREIGN KEY (user_id)
    REFERENCES "user" (id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT fk_video_invoice
    FOREIGN KEY (running_man_video_episode)
    REFERENCES running_man_video (episode)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);

CREATE TABLE payment (
  id VARCHAR(255) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  invoice_id UUID UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  CONSTRAINT fk_user_payment
    FOREIGN KEY (user_id)
    REFERENCES "user" (id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT fk_invoice_payment
    FOREIGN KEY (invoice_id)
    REFERENCES invoice (id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);