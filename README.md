# tabulae

API for media list management.

In the `/api/` folder:

- Running: `goapp serve`
- Deploying: `goapp deploy`
- Rollback: `appcfg.py rollback -A newsai-1166 -V 1 api/`

Indexes:

- Update: `appcfg.py update_indexes .`
- Delete: `appcfg.py vacuum_indexes .`
