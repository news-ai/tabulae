## NewsAI Docs

API

- `/api/users` (POST, GET)
- `/api/users/<id>` (PATCH, GET)
- `/api/agencies` (GET)
- `/api/agencies/<id>` (GET, PATCH)
- `/api/publications` (GET, POST)
- `/api/publications/<id>` (GET)
- `/api/contacts` (GET, POST, PATCH)
- `/api/contacts/<id>` (GET, PATCH)
- `/api/lists` (GET, POST)
- `/api/lists/<id>` (GET, PATCH)
- `/api/email/` (GET, POST, PATCH)
- `/api/emails/<id>` (GET, PATCH)

Login

- `/api/auth/google?next=http://newsai.org`
- `/api/auth`
- `/api/auth/registration`
- `/api/auth/logout`

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
CustomFields (customfields)

```json
{
    "name": "Pepsi",
    "customfields": ["Sex", "Cheese"]
}
```

PATCH

`/api/lists/<id>`

Name (name)
Contacts (contacts)
Archived (archived)

Contacts are array of contact Ids
Archived is a boolean

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
CustomFields (customfields)

WorksAt is an array of publication Ids

[0292309230923092, 0239209320932]

CustomFields is an array of key, value

```json
{
    "firstname": "Abhi",
    "customfields": [{
      "name": "Sex",
      "value": "Female"
    },{
        "name": "Cheese",
        "value": "YES!"
    }]
}
```

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

Email

POST

`http://localhost:8080/api/emails`

The `body` has to be on one line. No new lines. JSON doesn't allow that.

```json
{
"to": "abhia@nyu.edu",
"subject": "Beach?",
"body": "<!DOCTYPE html><html><head><title></title></head><body><p>Hi</p></body></html>"
}
```

BATCH POST (same as others)

PATCH (same as others)

BATCH PATCH (same as others)

Send email

GET request to

`http://localhost:8080/api/emails/<id>/send`

Login forward

`https://tabulae.newsai.org/api/auth/google?next=http://newsai.org`

Logout

`https://tabulae.newsai.org/api/auth/logout`
