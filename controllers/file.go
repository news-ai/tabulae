package controllers

import (
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/models"
)

// Generates a new key for the data to be stored on App Engine
func (f *models.File) key(c appengine.Context) *datastore.Key {
	if f.Id == 0 {
		f.Created = time.Now()
		return datastore.NewIncompleteKey(c, "File", nil)
	}
	return datastore.NewKey(c, "File", "", f.Id, nil)
}

/*
* Get methods
 */

func getFile(c appengine.Context, r *http.Request, id int64) (models.File, error) {
	// Get the File by id
	files := []models.File{}
	fileId := datastore.NewKey(c, "File", "", id, nil)
	ks, err := datastore.NewQuery("File").Filter("__key__ =", fileId).GetAll(c, &files)
	if err != nil {
		return models.File{}, err
	}
	if len(files) > 0 {
		files[0].Id = ks[0].IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.File{}, errors.New("Could not get user")
		}
		if files[0].CreatedBy != user.Id {
			return models.File{}, errors.New("Forbidden")
		}

		return files[0], nil
	}
	return models.File{}, errors.New("No file by this id")
}

/*
* Create methods
 */

func (f *models.File) create(c appengine.Context, r *http.Request) (*models.File, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return f, err
	}

	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	_, err = f.save(c)
	return f, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (f *models.File) save(c appengine.Context) (*models.File, error) {
	// Update the Updated time
	f.Updated = time.Now()

	k, err := datastore.Put(c, f.key(c), f)
	if err != nil {
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single file by the user
func GetFiles(c appengine.Context, r *http.Request) ([]models.File, error) {
	files := []models.File{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.File{}, err
	}

	ks, err := datastore.NewQuery("File").Filter("CreatedBy =", user.Id).GetAll(c, &files)
	if err != nil {
		return []models.File{}, err
	}

	for i := 0; i < len(files); i++ {
		files[i].Id = ks[i].IntID()
	}
	return files, nil
}

func GetFile(c appengine.Context, r *http.Request, id string) (models.File, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return models.File{}, err
	}

	file, err := getFile(c, r, currentId)
	if err != nil {
		return models.File{}, err
	}
	return file, nil
}

/*
* Create methods
 */

func CreateFile(r *http.Request, fileName string, listid string, createdby string) (models.File, error) {
	// Since upload.go uses a different appengine package
	c := appengine.NewContext(r)

	// Convert listId and createdById from string to int64
	listId, err := StringIdToInt(listid)
	if err != nil {
		return models.File{}, err
	}
	createdBy, err := StringIdToInt(createdby)
	if err != nil {
		return models.File{}, err
	}

	// Initialize file
	file := models.File{}
	file.FileName = fileName
	file.ListId = listId
	file.CreatedBy = createdBy

	// Create file
	_, err = file.create(c, r)
	if err != nil {
		return models.File{}, err
	}

	return file, nil
}

/*
* XLSX -> API methods
 */
