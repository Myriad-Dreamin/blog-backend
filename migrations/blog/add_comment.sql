-- https://stackoverflow.com/questions/4253804/how-do-i-add-a-new-column-in-between-two-columns
ALTER TABLE comments ADD COLUMN reply_to TEXT;
-- no I don't need it.
ALTER TABLE comments DROP COLUMN reply_to TEXT;
