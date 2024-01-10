CREATE TABLE breachdata (
   range text NOT NULL,
   key TEXT NOT NULL UNIQUE,
   count integer,
   unique(range, key)
);
