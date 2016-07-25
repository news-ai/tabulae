# tabulae

API for media monitoring

In the `/api/` folder:

- Running: `goapp serve`
- Deploying: `goapp deploy`
- Rollback: `appcfg.py rollback -A newsai-1166 -V 1 api/`

Progress:

- User (~~[GET](http://tabulae.newsai.org/api/users)~~, ~~PATCH~~)
- Publication (~~[GET](http://tabulae.newsai.org/api/publications)~~, ~~POST~~, PATCH, DELETE, BATCH POST)
- Lists (~~[GET](http://tabulae.newsai.org/api/lists)~~, ~~POST~~, ~~PATCH~~, DELETE)
- Agency (~~[GET](http://tabulae.newsai.org/api/agencies)~~, PATCH, DELETE)
- Contact (~~[GET](http://localhost:8080/api/contacts)~~, ~~POST~~, ~~PATCH~~, DELETE, ~~BATCH PATCH~~, ~~BATCH POST~~)

Later:

- Sync contacts through different lists
- If agency exists and new user signs up send everyone in that agency an email telling them that their friend X has signed up
