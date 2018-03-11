USE go_is_awesome;
ALTER TABLE Users ADD COLUMN Permissions int DEFAULT 10;
UPDATE Users
  SET Permissions = 1
  WHERE UserName = "Popeye";
