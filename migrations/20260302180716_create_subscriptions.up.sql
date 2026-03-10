-- create table
CREATE TABLE IF NOT EXISTS subscriptions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  service TEXT NOT NULL,
  price BIGINT NOT NULL,
  user_id UUID NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE
);

-- create index
CREATE INDEX IF NOT EXISTS ind_subscriptions_user_id ON subscriptions(user_id);