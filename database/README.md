## Script

```sql
CREATE DATABASE rust_postgres_server;

CREATE TABLE json_table (
   ID serial NOT NULL PRIMARY KEY,
   email VARCHAR NOT NULL,
   input JSONB NOT NULL,
   tags INTEGER[] DEFAULT '{}'
);
```