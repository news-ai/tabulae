Endpoints

- `/api/users`

    - GET -> Admin only

- `/api/users/<id>` (PATCH, GET)

    - GET "me" -> Current user
    - GET "id" -> Admin only
    - PATCH "me" -> Current user
    - PATCH "id" -> Admin only

- `/api/agencies` (GET)

    - GET -> Admin only

- `/api/agencies/<id>` (GET, PATCH)

    - GET -> Anyone logged in

- `/api/publications` (GET, POST)

    - GET -> Admin only
    - GET with filter ?name= -> Anyone logged in
    - POST -> Anyone logged in

- `/api/publications/<id>` (GET)

    - GET -> Anyone logged in

- `/api/contacts` (GET, POST, PATCH)

    - GET -> returns back information about contacts of yours
    - POST -> Creates a contact under you
    - PATCH -> Anyone logged in can update anyone's contact (FIX)

- `/api/contacts/<id>` (GET, PATCH)

    - GET -> Anyone logged in can get any contact (FIX)
    - PATCH -> Anyone logged in can update anyone's contact (FIX)

- `/api/lists` (GET, POST)

    - GET -> returns back information about your own contact lists
    - POST -> Creates a media list under you

- `/api/lists/<id>` (GET, PATCH)

    - GET -> Does not allow you to get someone else media list
    - PATCH -> Does not allow you to update your own media list

- `/api/email/` (GET, POST, PATCH)

    - GET -> returns back all information about your own emails
    - POST -> Create a new email
    - PATCH -> Anyone logged in can update anyone's email (FIX)

- `/api/emails/<id>` (GET, PATCH)

    - GET -> Anyone can get anyone's email (FIX)
    - PATCH -> Anyone logged in can update anyone's email (FIX)
