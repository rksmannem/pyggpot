1. schema changes:
   a. make sure that there is only one recode exists in `coin` table with same `pot_id` and `denomination`
   which will reduce complex logic required in removing the coins.
   b. another thing to consider is to remove corresponding records from `pot` table when  `coin` table when
   a pot is deleted from `pot` table.
   
```schema should look like:
CREATE TABLE IF NOT EXISTS coin (
id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
denomination integer NOT NULL,
coin_count integer NOT NULL,
pot_id integer NOT NULL,
FOREIGN KEY (pot_id) REFERENCES pot(id) ON DELETE CASCADE,
UNIQUE (pot_id, denomination)
)
```

2. it will be good if the request payload specifies the
   kind of coins and number of coins per each kind to remove, which will
   reduces complexity involved in identifying the random coins to remove.

3.  we can improve the performance by spinning a separate go routine for each coin kind if the request 
    is modified as mentioned in 2.
    