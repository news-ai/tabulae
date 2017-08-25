package controllers

import (
	"errors"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"

	"github.com/news-ai/api/controllers"
	"github.com/news-ai/api/models"

	"github.com/news-ai/tabulae/sync"
)

func RegisterUser(r *http.Request, user models.User) (models.User, bool, error) {
	c := appengine.NewContext(r)
	existingUser, err := controllers.GetUserByEmail(c, user.Email)

	if err != nil {
		// Validation if the email is null
		if user.Email == "" {
			noEmailErr := errors.New("User does have an email")
			log.Errorf(c, "%v", noEmailErr)
			log.Errorf(c, "%v", user)
			return models.User{}, false, noEmailErr
		}

		// Add the user to datastore
		_, err = user.Create(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return user, false, err
		}

		sync.ResourceSync(r, user.Id, "User", "create")

		// Set the user
		gcontext.Set(r, "user", user)
		controllers.Update(c, r, &user)

		// Create a sample media list for the user
		_, _, err = CreateSampleMediaList(c, r, user)
		if err != nil {
			log.Errorf(c, "%v", err)
		}
		return user, true, nil
	}

	if user.RefreshToken != "" {
		existingUser.RefreshToken = user.RefreshToken
	}

	if !existingUser.Gmail {
		existingUser.TokenType = user.TokenType
		existingUser.GoogleExpiresIn = user.GoogleExpiresIn
		existingUser.Gmail = user.Gmail
		existingUser.GoogleId = user.GoogleId
		existingUser.AccessToken = user.AccessToken
		existingUser.GoogleCode = user.GoogleCode
		existingUser.Save(c)
	}

	return existingUser, false, errors.New("User with the email already exists")
}
