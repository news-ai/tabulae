package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getFile(c context.Context, r *http.Request, id int64) (models.File, error) {
	// Get the File by id
	var file models.File
	fileId := datastore.NewKey(c, "File", "", id, nil)

	err := nds.Get(c, fileId, &file)
	if err != nil {
		return models.File{}, err
	}

	if !file.Created.IsZero() {
		file.Id = fileId.IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.File{}, errors.New("Could not get user")
		}
		if file.CreatedBy != user.Id && !user.IsAdmin {
			return models.File{}, errors.New("Forbidden")
		}

		return file, nil
	}
	return models.File{}, errors.New("No file by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single file by the user
func GetFiles(c context.Context, r *http.Request) ([]models.File, error) {
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

func GetFile(c context.Context, r *http.Request, id string) (models.File, error) {
	// Get the details of the current user
	currentId, err := utils.StringIdToInt(id)
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
	listId, err := utils.StringIdToInt(listid)
	if err != nil {
		return models.File{}, err
	}
	createdBy, err := utils.StringIdToInt(createdby)
	if err != nil {
		return models.File{}, err
	}

	// Initialize file
	file := models.File{}
	file.FileName = fileName
	file.ListId = listId
	file.CreatedBy = createdBy
	file.FileExists = true

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return file, err
	}

	// Create file
	_, err = file.Create(c, r, currentUser)
	if err != nil {
		return models.File{}, err
	}

	// Attach the fileId to the media list associated to it
	mediaList, err := GetMediaList(c, r, listid)
	if err != nil {
		return models.File{}, err
	}
	mediaList.FileUpload = file.Id
	mediaList.Save(c)

	return file, nil
}

/*
* XLSX -> API methods
 */
