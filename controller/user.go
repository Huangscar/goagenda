package controller

import (
	"fmt"

	"github.com/MegaShow/goagenda/lib/log"
)

type UserCtrl interface {
	UserDelete()
	UserList()
	UserSet()
}

func (c *Controller) UserDelete() {
	password, _ := c.Ctx.GetSecretString("password")
	userName, _ := c.Ctx.GetString("username")
	verifyUser(userName)
	verifyPassword(password)
	verifyEmptyArgs(c.Args)

	currentUser := c.Ctx.User.Get()
	if currentUser == "" {
		fmt.Println("you should login")
		return
	}

	if currentUser != userName {
		log.Error("you are already logged in with user '" + currentUser + "', please logout first")
	}

	err := c.Srv.User().DeleteUser(currentUser, password)
	if err != nil {
		log.Error(err.Error())
	}
	c.Ctx.User.Set("")
	log.Info("Delete account successfully")
}

func (c *Controller) UserList() {
	// TODO
	currentUser := c.Ctx.User.Get()
	if currentUser == "" {
		fmt.Println("you should login")
		return
	}

	userName, setN := c.Ctx.GetString("user")
	if !setN {
		fmt.Println(c.Srv.User().GetAllUsers())
	} else {
		verifyUser(userName)
		verifyEmptyArgs(c.Args)
		userDetail, err := c.Srv.User().GetUserDetail(userName)
		if err != nil {
			log.Error(err.Error())
		}
		fmt.Println(userDetail)
	}
}

func (c *Controller) UserSet() {
	password, setP := c.Ctx.GetSecretString("password")
	email, setE := c.Ctx.GetString("email")
	telephone, setT := c.Ctx.GetString("telephone")

	if setP && password == "" {
		log.Error("password empty")
	}
	verifyPassword(password)
	verifyEmail(email)
	verifyTelephone(telephone)
	verifyEmptyArgs(c.Args)

	log.Verbose("check status")
	currentUser := c.Ctx.User.Get()
	if currentUser == "" {
		fmt.Println("you should login")
		return
	}

	if !setP && !setE && !setT {
		fmt.Println("set nothing")
		return
	}
	err := c.Srv.User().Set(currentUser, password, setP, email, setE, telephone, setT)
	if err != nil {
		log.Error(err.Error())
	}
	log.Info("set user successfully")
}

func GetUserCtrl() UserCtrl {
	return &ctrl
}
