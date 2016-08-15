## NewsAI Docs

API

- `/api/users` (GET)
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

The next parameter is optional.

- `/api/auth/google?next=http://newsai.org`
- `/api/auth?next=http://newsai.org`
- `/api/auth/registration?next=http://newsai.org`
- `/api/auth/logout?next=http://newsai.org`

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

Upload

`/api/lists/<id>/upload`

```html
<!DOCTYPE html>
<html><body>
<form action="https://tabulae.newsai.org/api/lists/5453796448665600/upload" method="POST" enctype="multipart/form-data">
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
```

Sample form for uploading a excel sheet to the upload list. When you do a POST it'll return

```
{
    "id": 5683127032741888,
    "createdby": 5749563331706880,
    "created": "2016-08-13T16:12:06.975990163Z",
    "updated": "2016-08-13T16:12:06.975990513Z",
    "filename": "5749563331706880-5722383033827328-69ehnQ8=-KSchloss-EditorList(version1).xlsx",
    "listid": 5722383033827328
}
```

Then you can navigate to `https://tabulae.newsai.org/api/files/5683127032741888`, which is the ID you want.

Then to get the headers you can do a GET request on `https://tabulae.newsai.org/api/files/5683127032741888/headers`.

To set the headers you can do a POST request on `https://tabulae.newsai.org/api/files/5683127032741888/headers`.

POST with data like this:

```json
{"order": ["", ""]}
```

"ignore_column" ignores the column

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
