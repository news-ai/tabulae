## NewsAI Docs

- `/api`
- `/api/users` (POST, GET)
- `/api/users/<id>` (PATCH, GET)
- `/api/agencies` (GET)
- `/api/agencies/<id>` (PATCH, GET)
- `/api/publications` (GET, POST)
- `/api/publications/<id>` (GET)
- `/api/contacts` (GET, POST, PATCH)
- `/api/contacts/<id>` (GET)
- `/api/lists` (GET, POST)
- `/api/lists/<id>` (GET, PATCH)

----

Get user

`/api/users/me`

PATCH user

`/api/users/me`

FirstName (firstname)
LastName (lastname)
WorksAt (worksat) —> ID of agencies

WorksAt is an array of agency Ids


[0292309230923092, 0239209320932]
—

Create publication

POST

`/api/publications`

Name (name)
URL (url)

—

Create media lists

POST

`/api/lists`

Name (name)
Contacts (contacts)

PATCH

`/api/lists/<id>`

Name (name)
Contacts (contacts)

Contacts are array of contact Ids

same as WorksAt.

—

Create contact

POST

`/api/contacts`

FirstName (firstname)
LastName (lastname)
Email (email)
Linkedin (linkedin)
Instagram (instagram)
Twitter (twitter)
WorksAt (worksat)

WorksAt is an array of publication Ids

[0292309230923092, 0239209320932]

BATCH

```
[{
    "firstname": "SDSDSDS",
    "lastname": "gg"
},
{
    "firstname": "SDSDSDS2",
    "lastname": "gg"
}]
```

PATCH

`http://localhost:8080/api/contacts/4709895496531968`

```
{
    "lastname": "Cake",
    "email": "cheese@cake.com"
}
```

PATCH BATCH

`http://localhost:8080/api/contacts/`

```
[{
    "id": 4709895496531968,
    "firstname": "Agarwal",
    "lastname": "AAX",
    "email": ""
}, {
    "id": 5272845449953280,
    "firstname": "Abhi",
    "lastname": "SDS"
}]
```

the ids have to be int not string.

---

Login forward

http://tabulae.newsai.org/api/auth/google?next=http://newsai.org
