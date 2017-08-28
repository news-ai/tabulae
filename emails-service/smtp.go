package main

func func_name() {
	val, err := email.MarkSent(c, "")
	if err != nil {
		log.Errorf(c, "%v", err)
		return *val, nil, err
	}

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		emailBody, err := emails.GenerateEmail(r, user, email, files)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		emailSetting, err := getEmailSetting(c, r, user.EmailSetting)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		SMTPPassword := string(user.SMTPPassword[:])

		contextWithTimeout, _ := context.WithTimeout(c, time.Second*30)
		client := urlfetch.Client(contextWithTimeout)
		getUrl := "https://tabulae-smtp.newsai.org/send"

		sendEmailRequest := models.SMTPEmailSettings{}
		sendEmailRequest.Servername = emailSetting.SMTPServer + ":" + strconv.Itoa(emailSetting.SMTPPortSSL)
		sendEmailRequest.EmailUser = user.SMTPUsername
		sendEmailRequest.EmailPassword = SMTPPassword
		sendEmailRequest.To = email.To
		sendEmailRequest.Subject = email.Subject
		sendEmailRequest.Body = emailBody

		SendEmailRequest, err := json.Marshal(sendEmailRequest)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}
		log.Infof(c, "%v", string(SendEmailRequest))
		sendEmailQuery := bytes.NewReader(SendEmailRequest)

		req, _ := http.NewRequest("POST", getUrl, sendEmailQuery)

		resp, err := client.Do(req)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var verifyResponse SMTPEmailResponse
		err = decoder.Decode(&verifyResponse)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		log.Infof(c, "%v", verifyResponse)

		if verifyResponse.Status {
			val, err = email.MarkDelivered(c)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}
			return *val, nil, nil
		}

		return *val, nil, errors.New(verifyResponse.Error)
	}

	return *val, nil, nil
}
