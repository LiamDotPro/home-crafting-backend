package main

import (
	"fmt"
	"github.com/LiamDotPro/Go-Multitenancy/helpers"
	"github.com/LiamDotPro/Go-Multitenancy/params"
	"github.com/gin-gonic/gin"
	"github.com/wader/gormstore"
	"net/http"
	"time"
)

type CustomerProfile struct {
	LoginAttempts        map[string]*loginAttempt // Key is used email address
	LastLoginAttemptTime time.Time
	AuthorizedTime       time.Time
	UserId               uint
	Authorized           uint
}

type HostProfile struct {
	LoginAttempts        map[string]*loginAttempt // Key is used email address
	LastLoginAttemptTime time.Time
	AuthorizedTime       time.Time
	UserId               uint
	Authorized           uint
}

type loginAttempt struct {
	LastLoginAttemptTime time.Time
	LoginAttempts        uint
}

func newHostProfile() HostProfile {
	h := HostProfile{}
	h.LoginAttempts = make(map[string]*loginAttempt)
	h.Authorized = 0
	return h
}

func NewCustomerProfile() CustomerProfile {
	c := CustomerProfile{}
	c.LoginAttempts = make(map[string]*loginAttempt)
	c.Authorized = 0
	return c
}

// Checks if a user is logged in with a session to the master dashboard
func HandleMasterLoginAttempt(Store *gormstore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Try and get a session.
		sessionValues, err := Store.Get(c.Request, "connect.s.id")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
			c.Abort()
			return
		}

		// Check to see if the user is already authorized..
		if sessionValues.ID != "" {

			p := sessionValues.Values["host"].(HostProfile)

			if p.Authorized == 1 {
				c.JSON(http.StatusOK, gin.H{
					"outcome": "Already Authorized",
					"message": "user already authorized with application.",
				})
				c.Abort()
				return
			}
		}

		// Check our parameters out.
		var json params.LoginParams

		// Abort if we don't have the correct variables to begin with.
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email or Password provided are incorrect, please try again."})
			fmt.Println("Can't bind request variables for login")
			c.Abort()
			return
		}

		if !helpers.ValidateEmail(json.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email or Password provided are incorrect, please try again."})
			fmt.Println("Email is not in a valid format.")
			c.Abort()
			return
		}

		// Validate the password being sent.
		if len(json.Password) <= 7 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The specified password was to short, must be longer than 8 characters."})
			c.Abort()
			return
		}

		// Validate the password contains at least one letter and capital
		if !helpers.ContainsCapitalLetter(json.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The specified password does not contain a capital letter."})
			c.Abort()
			return
		}

		// Make sure the password contains at least one special character.
		if !helpers.ContainsSpecialCharacter(json.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The password must contain at least one special character."})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
			c.Abort()
			return
		}

		c.Set("bindedJson", json)

		// Check to see if a new session is found.
		if sessionValues.ID == "" {
			// Setup new session with empty profiles.
			session, err := Store.New(c.Request, "connect.s.id")

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
				c.Abort()
				return
			}

			// Host profile requires little setup
			session.Values["host"] = newHostProfile()

			// Add the entry record to hostProfile
			session.Values["host"].(HostProfile).LoginAttempts[json.Email] = &loginAttempt{LoginAttempts: 1, LastLoginAttemptTime: time.Now().UTC()}

			// Client profile requires no setup
			session.Values["customer"] = NewCustomerProfile()

			// Set the session back to the handler for use.
			c.Set("session", session)
			return
		} else {
			// Profile was already found
			h := sessionValues.Values["host"].(HostProfile)

			// Check if the email used is already in our login attempts.
			loginAttemptsFound, found := h.LoginAttempts[json.Email]

			if !found {
				// email has not been used to login add a new entry
				h.LoginAttempts[json.Email] = &loginAttempt{LoginAttempts: 1, LastLoginAttemptTime: time.Now().UTC()}

				// Set the session back to the handler for use.
				c.Set("session", sessionValues)
				return
			}

			// Check to see if login attempts exceeds 3 attempts
			if found && loginAttemptsFound.LoginAttempts > 2 {
				// Check to see if last login attempt was over half an hour ago
				if time.Now().Sub(loginAttemptsFound.LastLoginAttemptTime).Minutes() > 30 {
					// reset login attempts to have 2 more.
					loginAttemptsFound.LoginAttempts = 1
					loginAttemptsFound.LastLoginAttemptTime = time.Now().UTC()

					// Set the session back to the handler for use.
					c.Set("session", sessionValues)
					return
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "You have been locked out for too many attempts to login..", "status": "locked out", "timeLeft": 30 - time.Now().Sub(loginAttemptsFound.LastLoginAttemptTime).Minutes()})
					c.Abort()
					return
				}
			}

			if found && loginAttemptsFound.LoginAttempts <= 2 {
				// increase login attempt count
				loginAttemptsFound.LoginAttempts++
				// replace last attempt date
				loginAttemptsFound.LastLoginAttemptTime = time.Now().UTC()

				// Set the session back to the handler for use.
				c.Set("session", sessionValues)
				return
			}

		}

	}

}

// Checks if a user is logged in with a session to the master dashboard
func HandleCustomerLogin(Store *gormstore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Try and get a session.
		sessionValues, err := Store.Get(c.Request, "connect.s.id")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
			c.Abort()
			return
		}

		// Check to see if the user is already authorized..
		if sessionValues.ID != "" {

			p := sessionValues.Values["client"].(CustomerProfile)

			if p.Authorized == 1 {
				c.JSON(http.StatusOK, gin.H{
					"outcome": "Already Authorized",
					"message": "user already authorized with application.",
				})
				c.Abort()
				return
			}
		}

		// Check our parameters out.
		var json params.LoginParams

		// Abort if we don't have the correct variables to begin with.
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email or Password provided are incorrect, please try again."})
			fmt.Println("Can't bind request variables for login")
			c.Abort()
			return
		}

		if !helpers.ValidateEmail(json.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email or Password provided are incorrect, please try again."})
			fmt.Println("Email is not in a valid format.")
			c.Abort()
			return
		}

		// Validate the password being sent.
		if len(json.Password) <= 7 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The specified password was to short, must be longer than 8 characters."})
			c.Abort()
			return
		}

		// Validate the password contains at least one letter and capital
		if !helpers.ContainsCapitalLetter(json.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The specified password does not contain a capital letter."})
			c.Abort()
			return
		}

		// Make sure the password contains at least one special character.
		if !helpers.ContainsSpecialCharacter(json.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The password must contain at least one special character."})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
			c.Abort()
			return
		}

		c.Set("bindedJson", json)

		// Check to see if a new session is found.
		if sessionValues.ID == "" {
			// Setup new session with empty profiles.
			session, err := Store.New(c.Request, "connect.s.id")

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong.."})
				c.Abort()
				return
			}

			// Host profile requires little setup
			session.Values["host"] = newHostProfile()

			// Client profile requires no setup
			session.Values["customer"] = NewCustomerProfile()

			// Add the entry record to hostProfile
			session.Values["customer"].(CustomerProfile).LoginAttempts[json.Email] = &loginAttempt{LoginAttempts: 1, LastLoginAttemptTime: time.Now().UTC()}

			// Set the session back to the handler for use.
			c.Set("session", session)
			return
		} else {
			// Profile was already found
			h := sessionValues.Values["customer"].(CustomerProfile)

			// Check if the email used is already in our login attempts.
			loginAttemptsFound, found := h.LoginAttempts[json.Email]

			if !found {
				// email has not been used to login add a new entry
				h.LoginAttempts[json.Email] = &loginAttempt{LoginAttempts: 1, LastLoginAttemptTime: time.Now().UTC()}

				// Set the session back to the handler for use.
				c.Set("session", sessionValues)
				return
			}

			// Check to see if login attempts exceeds 3 attempts
			if found && loginAttemptsFound.LoginAttempts > 2 {
				// Check to see if last login attempt was over half an hour ago
				if time.Now().Sub(loginAttemptsFound.LastLoginAttemptTime).Minutes() > 30 {
					// reset login attempts to have 2 more.
					loginAttemptsFound.LoginAttempts = 1
					loginAttemptsFound.LastLoginAttemptTime = time.Now().UTC()

					// Set the session back to the handler for use.
					c.Set("session", sessionValues)
					return
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "You have been locked out for too many attempts to login..", "status": "locked out", "timeLeft": 30 - time.Now().Sub(loginAttemptsFound.LastLoginAttemptTime).Minutes()})
					c.Abort()
					return
				}
			}

			if found && loginAttemptsFound.LoginAttempts <= 2 {
				// increase login attempt count
				loginAttemptsFound.LoginAttempts++
				// replace last attempt date
				loginAttemptsFound.LastLoginAttemptTime = time.Now().UTC()

				// Set the session back to the handler for use.
				c.Set("session", sessionValues)
				return
			}

		}

	}

}
