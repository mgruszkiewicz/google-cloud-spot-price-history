# Google Cloud Spot Price History

build sqlite database with historical spot instance prices based on [google-cloud-pricing-cost-calculator](https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator) 

## How it works?

First, we clone the google-cloud-pricing-cost-calculator repo, then we extract all versions of file `pricing.yml`.  
That file contains pricing of all GCE products, so it might be usefull also for other stuff. Then we process all yaml files from output and populate the values in sqlite3 database - we don't need anything more fancy.  
Now we can query using SQL the price history.
```sql
sqlite> select * from pricing_history where region_name="europe-west1" AND machine_type="t2d-standard-4" ORDER BY updated ASC;
9117|t2d-standard-4|europe-west1|0.185892|0.049464|1683528167|2023-05-08 06:42:47+00:00
13666|t2d-standard-4|europe-west1|0.185892|0.049464|1683634028|2023-05-09 12:07:08+00:00
[...]
1216842|t2d-standard-4|europe-west1|0.185892|0.041112|1747495189|2025-05-17 15:19:49+00:00
1229344|t2d-standard-4|europe-west1|0.185892|0.041112|1747886484|2025-05-22 04:01:24+00:00
```

## How to build
To build this script, you need Golang 1.24+

```
make build
```

In `bin/` directory you will find the binary for your system.

## Ideas for future

* [ ] WebUI to visualize the changes?
* [ ] Automatic sqlite3 build in gh actions on change


## Acknowledgements
* [Cyclenerd](https://github.com/Cyclenerd) - creator of [google-cloud-pricing-cost-calculator](https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator) project