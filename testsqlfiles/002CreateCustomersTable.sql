USE go_is_awesome;
CREATE TABLE Customers (
    ID int NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (ID),
    CustomerName varchar(255),
    ContactName varchar(255),
    Address varchar(255),
    City varchar(255),
    PostalCode varchar(255),
    Country varchar(255)
);

INSERT INTO Customers (
  CustomerName,
  ContactName,
  Address,
  City,
  PostalCode,
  Country
) VALUES (
  'Cardinal',
  'Tom B. Erichsen;',
  'Skagen 21',
  'Stavanger',
  '4006',
  'Norway'
), (
  'White Clover Markets',
  'Karl Jablonski',
  '305 - 14th Ave. S. Suite 3B',
  'Seattle',
  '98128',
  'USA'
),(
  'Wilman Kala',
	'Matti Karttunen',
  'Keskuskatu 45',
  'Helsinki',
  '21240',
  'Finland'
);
