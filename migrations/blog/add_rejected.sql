-- Distinguish comments that were explicitly rejected from comments still awaiting review.
ALTER TABLE comments ADD COLUMN rejected BOOLEAN NOT NULL DEFAULT FALSE;
