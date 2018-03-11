USE go_is_awesome;
CREATE TABLE Users (
    ID int NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (ID),
    UserName varchar(255),
    Email varchar(255)
);

INSERT INTO Users (
  UserName,
  Email
) VALUES (
  "Popeye",
  "popeye@olive.io"
), (
  "Chicken",
  "chicken@roads.net"
);
