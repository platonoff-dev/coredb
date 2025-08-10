# Structure of file and logic blocks


## File structure (in pages)

| Description                | Size |
|----------------------------|------|
| DB Header page             | 4096 |
| Freelist linked list pages | -    |
| Schema pages               | -    |
| Data pages                 | -    |

## Page structure

Structure of all pages except for the DB Header page:

| Description        | Size               |
|--------------------|--------------------|
| Page type          | 1                  |
| Free space offset  | 2                  |
| Page data          | All available space|
