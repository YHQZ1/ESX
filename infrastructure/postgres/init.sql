CREATE DATABASE order_gateway;
CREATE DATABASE participant_registry;
CREATE DATABASE risk_engine;
CREATE DATABASE matching_engine;
CREATE DATABASE clearing_house;
CREATE DATABASE market_data_feed;
CREATE DATABASE settlement_engine;
CREATE DATABASE ledger_service;

GRANT ALL PRIVILEGES ON DATABASE order_gateway TO esx;
GRANT ALL PRIVILEGES ON DATABASE participant_registry TO esx;
GRANT ALL PRIVILEGES ON DATABASE risk_engine TO esx;
GRANT ALL PRIVILEGES ON DATABASE matching_engine TO esx;
GRANT ALL PRIVILEGES ON DATABASE clearing_house TO esx;
GRANT ALL PRIVILEGES ON DATABASE market_data_feed TO esx;
GRANT ALL PRIVILEGES ON DATABASE settlement_engine TO esx;
GRANT ALL PRIVILEGES ON DATABASE ledger_service TO esx;