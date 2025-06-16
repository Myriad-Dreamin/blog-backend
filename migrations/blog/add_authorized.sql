-- https://stackoverflow.com/questions/4253804/how-do-i-add-a-new-column-in-between-two-columns
ALTER TABLE comments ADD COLUMN authorized BOOLEAN NOT NULL DEFAULT FALSE;
