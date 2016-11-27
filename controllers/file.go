package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
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
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	if !file.Created.IsZero() {
		file.Format(fileId, "files")

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.File{}, errors.New("Could not get user")
		}
		if file.CreatedBy != user.Id && !user.IsAdmin {
			return models.File{}, errors.New("Forbidden")
		}

		return file, nil
	}
	return models.File{}, errors.New("No file by this id")
}

func getFileUnauthorized(c context.Context, r *http.Request, id int64) (models.File, error) {
	// Get the File by id
	var file models.File
	fileId := datastore.NewKey(c, "File", "", id, nil)

	err := nds.Get(c, fileId, &file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	if !file.Created.IsZero() {
		file.Format(fileId, "files")
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
func GetFiles(c context.Context, r *http.Request) ([]models.File, interface{}, int, error) {
	files := []models.File{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.File{}, nil, 0, err
	}

	query := datastore.NewQuery("File").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.File{}, nil, 0, err
	}

	files = make([]models.File, len(ks))
	err = nds.GetMulti(c, ks, files)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.File{}, nil, 0, err
	}

	for i := 0; i < len(files); i++ {
		files[i].Format(ks[i], "files")
	}
	return files, nil, len(files), nil
}

func GetFile(c context.Context, r *http.Request, id string) (models.File, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, nil, err
	}

	file, err := getFile(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, nil, err
	}
	return file, nil, nil
}

func GetFileById(c context.Context, r *http.Request, id int64) (models.File, interface{}, error) {
	file, err := getFile(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, nil, err
	}
	return file, nil, nil
}

func GetFileByIdUnauthorized(c context.Context, r *http.Request, id int64) (models.File, interface{}, error) {
	file, err := getFileUnauthorized(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, nil, err
	}
	return file, nil, nil
}

func FilterFileByImported(c context.Context, r *http.Request) ([]models.File, error) {
	// Get a publication by the URL
	ks, err := datastore.NewQuery("File").Filter("Imported =", true).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.File{}, err
	}

	if len(ks) == 0 {
		return []models.File{}, errors.New("No files by the field Imported")
	}

	var files []models.File
	files = make([]models.File, len(ks))
	err = nds.GetMulti(c, ks, files)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.File{}, err
	}

	nonImageFiles := []models.File{}
	for i := 0; i < len(files); i++ {
		if files[i].Url == "" {
			files[i].Format(ks[i], "files")
			nonImageFiles = append(nonImageFiles, files[i])
		}
	}
	return nonImageFiles, nil
}

/*
* Create methods
 */

func CreateFile(r *http.Request, fileName string, listid string, createdby string) (models.File, error) {
	// Since upload.go uses a different appengine package
	c := appengine.NewContext(r)

	// Convert listId and createdById from string to int64
	listId, err := utilities.StringIdToInt(listid)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Errorf(c, "%v", err)
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
		log.Errorf(c, "%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	// Attach the fileId to the media list associated to it
	mediaList, _, err := GetMediaList(c, r, listid)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}
	mediaList.FileUpload = file.Id
	mediaList.Save(c)

	return file, nil
}

func CreateImageFile(r *http.Request, originalFilename string, fileName string, createdby string, bucket string) (models.File, error) {
	// Since upload.go uses a different appengine package
	c := appengine.NewContext(r)

	// Convert listId and createdById from string to int64
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	publicURL := "https://storage.googleapis.com/%s/%s"

	// Initialize file
	file := models.File{}
	file.OriginalName = originalFilename
	file.FileName = fileName
	file.CreatedBy = createdBy
	file.FileExists = true
	file.Url = fmt.Sprintf(publicURL, bucket, fileName)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	return file, nil
}

func CreateAttachmentFile(r *http.Request, originalFilename string, fileName string, emailid string, createdby string) (models.File, error) {
	// Since upload.go uses a different appengine package
	c := appengine.NewContext(r)

	// Convert listId and createdById from string to int64
	emailId, err := utilities.StringIdToInt(emailid)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	// Initialize file
	file := models.File{}
	file.OriginalName = originalFilename
	file.FileName = fileName
	file.EmailId = emailId
	file.CreatedBy = createdBy
	file.FileExists = true

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	// Attach attachment to email
	email, err := getEmail(c, r, emailId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.File{}, err
	}

	email.Attachments = append(email.Attachments, file.Id)
	email.Save(c)

	return file, nil
}

/*
* XLSX -> API methods
 */
