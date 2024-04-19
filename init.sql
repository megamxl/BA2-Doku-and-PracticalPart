CREATE TABLE packages (
                          id SERIAL PRIMARY KEY,
                          tracking_number VARCHAR(255) NOT NULL UNIQUE,
                          sender VARCHAR(255) NOT NULL,
                          recipient VARCHAR(255) NOT NULL,
                          origin_address VARCHAR(255) NOT NULL,
                          destination_address VARCHAR(255) NOT NULL,
                          weight int8 NOT NULL,
                          status VARCHAR(50) NOT NULL
);

CREATE SEQUENCE package_id_seq;

DROP TABLE packages;

INSERT INTO packages
(id, tracking_number, sender, recipient, origin_address, destination_address, weight, status)
VALUES
    (nextval('package_id_seq'), 'TN123456789', 'John Doe', 'Jane Smith', '123 Sender Lane, New York, NY', '456 Recipient St, Los Angeles, CA', 3, 'shipped');

INSERT INTO packages
( tracking_number, sender, recipient, origin_address, destination_address, weight, status)
VALUES
    ( 'TN987654321', 'Alice Johnson', 'Bob Brown', '789 Sender Road, Chicago, IL', '321 Recipient Ave, Houston, TX', 5.75, 'in transit');

INSERT INTO packages
( tracking_number, sender, recipient, origin_address, destination_address, weight, status)
VALUES
    ( 'TN192837465', 'Charlie Kim', 'Dana Lee', '123 Sender Blvd, Miami, FL', '654 Recipient Path, Seattle, WA', 3.20, 'delivered');


SELECT * FROM packages;