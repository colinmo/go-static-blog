Title: "MySQL Learnings"
Tags: [code,mysql]
Created: "2021-03-31T17:11:15+1000"
Updated: "2021-03-31T17:11:15+1000"
Type: article
Status: live
Synopsis: "Upskilling my MySQL because I need a 6Gig database in my life created in Golang"
FeatureImage: /blog/media/2021/03/mysql-logo.svg
===

## Batching inserts

* [How do I batch sql statements with package database/sql](https://stackoverflow.com/questions/12486436/how-do-i-batch-sql-statements-with-package-database-sql#25192138)

Seriously, this one weird trick reduced insert time by a _tonne_. I really have to benchmark this at some point as it's a lifesaver.

## Explain/ Execution Plans

* [8.8 Understanding the Query Execution Plan](https://dev.mysql.com/doc/refman/8.0/en/execution-plan-information.html)

Speaking of poor performing queries...

## Running MySQL in the background

* [Stored procedure execution in background?](https://stackoverflow.com/questions/56198880/stored-procedure-execution-in-background)
* [Stored procedures](https://www.mysqltutorial.org/mysql-stored-procedure-tutorial.aspx)

With a 6Gig database there's a lot of things I need to kick off in the background for long run while the GoCode does other things. Building function+logic into a Procedure sounds sensible, and then when Go's prepped some databases calling a once-off-background procedure run makes sense.

```sql
CREATE EVENT run_info_now
ON SCHEDULE AT CURRENT_TIMESTAMP
DO CALL simbakda_sensus.info;
```

## Cursors

* [MySQL Cursor](https://www.mysqltutorial.org/mysql-cursor/)

Good MySQL cursor refresher.

## Triggers

... and how to disable them. Short version, you can't. Long version, make a variable and do it yourself.

* [How to disable triggers in MySQL?](https://stackoverflow.com/questions/13598155/how-to-disable-triggers-in-mysql)