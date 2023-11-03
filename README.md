# Eskiz-uz

Sign up or sign in and get your password from [eskiz.uz](https://eskiz.uz)

API Full documentation [here](https://documenter.getpostman.com/view/663428/RzfmES4z?version=latest)



### Example

```go
package main

import (
	"github.com/realtemirov/eskizuz"
)

func main() {
	eskiz, err := eskizuz.GetToken(&eskizuz.Auth{
		Email:    "your_email",
		Password: "your_sms_service_password",
	})
	if err != nil {
		panic(err)
	}

	sms := &eskizuz.SMS{
		MobilePhone: "998946992809",
		Message:      "test-message",
		From:         "go-eskiz-uz",
		CallbackURL:  "https://oxbox.udevs.io",
	}
	
    // Sending message
	result, err := eskiz.Send(sms)

    // Refresh token
    err = eskiz.RefreshToken()

    // User info
    user, err := eskiz.UserInfo()

    // get user limit
    result, err := eskiz.GetUserLimit()

}
```


# Contributing
If you get errors, please create an issue or pull request.

