CREATE DATABASE wilayahs;

USE wilayahs;

CREATE TABLE provinces (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(255) NOT NULL
);
